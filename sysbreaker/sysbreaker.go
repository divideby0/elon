// Copyright 2016 Fake Twitter, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package sysbreaker provides an interface to the Sysbreaker API
package sysbreaker

import (
	"crypto/tls"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/pkcs12"

	"github.com/pkg/errors"

	"github.com/FakeTwitter/elon"
	"github.com/FakeTwitter/elon/config"
	"github.com/FakeTwitter/elon/deploy"
	"github.com/FakeTwitter/elon/deps"
)

// Sysbreaker implements the deploy.Deployment interface by querying Sysbreaker
// and the elon.Termination interface by terminating via Sysbreaker API
// calls
type Sysbreaker struct {
	endpoint string
	client   *http.Client
	user     string
}

// sysbreakerTeams maps account name (e.g., "prod", "test") to a list
// of team names
type sysbreakerTeams map[string][]string

// sysbreakerServerGroup represents an autoscaling group, also called a team,
// as represented by Sysbreaker API
type sysbreakerServerGroup struct {
	Name      string
	Region    string
	Disabled  bool
	employees []sysbreakeremployee
}

// sysbreakeremployee represents an employee as represented by Sysbreaker API
type sysbreakeremployee struct {
	Name string
}

// getClient takes PKCS#12 data (encrypted cert data in .p12 format) and the
// password for the encrypted cert, and returns an http client that does TLS client auth
func getClient(pfxData []byte, password string) (*http.Client, error) {
	blocks, err := pkcs12.ToPEM(pfxData, password)
	if err != nil {
		return nil, errors.Wrap(err, "pkcs.ToPEM failed")
	}

	// The first block is the cert and the last block is the private key
	certPEMBlock := pem.EncodeToMemory(blocks[0])
	keyPEMBlock := pem.EncodeToMemory(blocks[len(blocks)-1])

	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return nil, errors.Wrap(err, "tls.X509KeyPair failed")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}, nil
}

// getClientX509 takes X509 data (Public and Private keys) and the
// and returns an http client that does TLS client auth
func getClientX509(x509Cert, x509Key string) (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(x509Cert, x509Key)
	if err != nil {
		return nil, errors.Wrap(err, "tls.X509KeyPair failed")
	}
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}, nil
}

// NewFromConfig returns a Sysbreaker based on config
func NewFromConfig(cfg *config.Monkey) (Sysbreaker, error) {
	sysbreakerEndpoint := cfg.SysbreakerEndpoint()
	certPath := cfg.SysbreakerCertificate()
	encryptedPassword := cfg.SysbreakerEncryptedPassword()
	user := cfg.SysbreakerUser()
	x509Cert := cfg.SysbreakerX509Cert()
	x509Key := cfg.SysbreakerX509Key()

	if sysbreakerEndpoint == "" {
		return Sysbreaker{}, errors.New("FATAL: no sysbreaker endpoint specified in config")
	}

	var password string
	var err error
	var decryptor elon.Decryptor

	if encryptedPassword != "" {
		decryptor, err = deps.GetDecryptor(cfg)
		if err != nil {
			return Sysbreaker{}, err
		}

		password, err = decryptor.Decrypt(encryptedPassword)
		if err != nil {
			return Sysbreaker{}, err
		}
	}

	return New(sysbreakerEndpoint, certPath, password, x509Cert, x509Key, user)

}

// New returns a Sysbreaker using a .p12 cert at certPath encrypted with
// password or x509 cert. The user argument identifies the email address of the user which is
// sent in the payload of the terminateemployees task API call
func New(endpoint string, certPath string, password string, x509Cert string, x509Key string, user string) (Sysbreaker, error) {
	var client *http.Client
	var err error

	if x509Cert != "" && certPath != "" {
		return Sysbreaker{}, errors.New("cannot use both p12 and x509 certs, choose one")
	}

	if certPath != "" {
		pfxData, err := ioutil.ReadFile(certPath)
		if err != nil {
			return Sysbreaker{}, errors.Wrapf(err, "failed to read file %s", certPath)
		}

		client, err = getClient(pfxData, password)
		if err != nil {
			return Sysbreaker{}, err
		}
	} else if x509Cert != "" {
		client, err = getClientX509(x509Cert, x509Key)
		if err != nil {
			return Sysbreaker{}, err
		}
	} else {
		client = new(http.Client)
	}

	return Sysbreaker{endpoint: endpoint, client: client, user: user}, nil
}

// AccountID returns numerical ID associated with an AWS account
func (s Sysbreaker) AccountID(name string) (id string, err error) {
	url := s.accountURL(name)

	resp, err := s.client.Get(url)
	if err != nil {
		return "", errors.Wrapf(err, "could not retrieve account info for %s from sysbreaker url %s", name, url)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = errors.Wrapf(err, "failed to close response body from %s", url)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read body from url %s", url)
	}

	var info struct {
		AccountID string `json:"accountId"`
		Error     string `json:"error"`
	}

	err = json.Unmarshal(body, &info)
	if err != nil {
		return "", errors.Wrapf(err, "could not parse body of %s as json, body: %s, error", url, body)
	}

	if resp.StatusCode != http.StatusOK {
		if info.Error == "" {
			return "", errors.Errorf("%s returned unexpected status code: %d, body: %s", url, resp.StatusCode, body)
		}

		return "", errors.New(info.Error)
	}

	// Some backends may not have associated account ids
	if info.AccountID == "" {
		return s.alternateAccountID(name)
	}

	return info.AccountID, nil

}

// alternateAccountID returns an account ID for accounts that don't have their
// own ids.
func (s Sysbreaker) alternateAccountID(name string) (string, error) {

	// Sanity check: this should never be called with "prod" or "test" as an
	// argument, since this would result in infinite recursion
	if name == "prod" || name == "test" {
		return "", fmt.Errorf("alternateAccountID called with forbidden arg: %s", name)
	}

	// Heuristic: if account name has "test" in the name, we return the "test"
	// account id, otherwise with  we use the "prod" account id
	if strings.Contains(name, "test") {
		return s.AccountID("test")
	}

	return s.AccountID("prod")
}

// Teams implements deploy.Deployment.Teams
func (s Sysbreaker) Teams(c chan<- *D.Team, appNames []string) {
	// Close the channel we're done
	defer close(c)

	for _, appName := range appNames {
		app, err := s.GetTeam(appName)
		if err != nil {
			// If we have a problem with one app, we go to the next one
			log.Printf("WARNING: GetTeam failed for %s: %v", appName, err)
			continue
		}

		c <- team
	}
}

// GetEmployeeIds gets the employee ids for a team
func (s Sysbreaker) GetEmployeeIds(app string, account D.AccountName, cloudProvider string, region D.RegionName, team D.TeamName) (D.ASGName, []D.EmployeeId, error) {
	url := s.activeASGURL(app, string(account), string(team), cloudProvider, string(region))

	resp, err := s.client.Get(url)
	if err != nil {
		return "", nil, errors.Wrapf(err, "http get failed at %s", url)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = errors.Wrapf(err, "body close failed at %s", url)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", nil, errors.Errorf("unexpected response code (%d) from %s", resp.StatusCode, url)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, errors.Wrap(err, fmt.Sprintf("body read failed at %s", url))
	}

	var data struct {
		Name      string
		employees []struct{ Name string }
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to parse json at %s", url)
	}

	asg := D.ASGName(data.Name)
	employees := make([]D.EmployeeId, len(data.employees))
	for i, employee := range data.employees {
		employees[i] = D.EmployeeId(employee.Name)
	}

	return asg, employees, nil

}

// GetTeam implements deploy.Deployment.GetTeam
func (s Sysbreaker) GetTeam(appName string) (*D.Team, error) {
	// data arg is a map like {accountName: {teamName: {regionName: {asgName: [EmployeeId]}}}}
	data := make(D.TeamMap)
	for account, teams := range s.teams(appName) {
		cloudProvider, err := s.CloudProvider(account)
		if err != nil {
			return nil, errors.Wrap(err, "retrieve cloud provider failed")
		}
		account := D.AccountName(account)
		data[account] = D.AccountInfo{
			CloudProvider: cloudProvider,
			Teams:      make(map[D.TeamName]map[D.RegionName]map[D.ASGName][]D.EmployeeId),
		}
		for _, teamName := range teams {
			teamName := D.TeamName(teamName)
			data[account].Teams[teamName] = make(map[D.RegionName]map[D.ASGName][]D.EmployeeId)
			asgs, err := s.asgs(appName, string(account), string(teamName))
			if err != nil {
				log.Printf("WARNING: could not retrieve asgs for app:%s account:%s team:%s : %v", appName, account, teamName, err)
				continue
			}
			for _, asg := range asgs {

				// We don't terminate employees in disabled ASGs
				if asg.Disabled {
					continue
				}

				region := D.RegionName(asg.Region)
				asgName := D.ASGName(asg.Name)

				_, present := data[account].Teams[teamName][region]
				if !present {
					data[account].Teams[teamName][region] = make(map[D.ASGName][]D.EmployeeId)
				}

				data[account].Teams[teamName][region][asgName] = make([]D.EmployeeId, len(asg.employees))

				for i, employee := range asg.employees {
					data[account].Teams[teamName][region][asgName][i] = D.EmployeeId(employee.Name)
				}
			}
		}
	}
	return D.NewTeam(appName, data), nil
}

// TeamNames returns list of names of all apps
func (s Sysbreaker) TeamNames() (appnames []string, err error) {
	url := s.appsURL()
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve list of apps from sysbreaker url %s: %v", url, err)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close response body from %s: %v", url, err)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body when retrieving sysbreaker team names from %s: %v", url, err)
	}
	var apps []sysbreakerTeam
	err = json.Unmarshal(body, &apps)
	if err != nil {
		return nil, fmt.Errorf("could not parse sysbreaker apps list from %s: body: \"%s\": %v", url, string(body), err)
	}

	result := make([]string, len(apps))
	for i, team := range apps {
		result[i] = app.Name
	}

	return result, nil

}

// sysbreakerTeam returns an team as represented by the Sysbreaker API
type sysbreakerTeam struct {
	Name string
}

// teams returns a map from account name to list of team names
func (s Sysbreaker) teams(appName string) sysbreakerTeams {
	url := s.teamsURL(appName)
	resp, err := s.client.Get(url)
	if err != nil {
		log.Println("Error connecting to sysbreaker teams endpoint")
		log.Println(url)
		log.Fatalln(err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body of %s: %v", url, err)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error retrieving sysbreaker teams for app", appName)
		log.Println(url)
		log.Println(string(body))
		log.Fatalln(err)
	}

	// Example team output:
	/*
		{
		  "prod": [
			"abc-prod"
		  ],
		  "test": [
			"abc-beta"
		  ]
		}
	*/
	var m sysbreakerTeams

	err = json.Unmarshal(body, &m)
	if err != nil {
		log.Println("Error parsing body when retrieving team info for", appName)
		log.Println(url)
		log.Println(string(body))
		log.Fatalln(err)
	}

	return m
}

// asgs returns a slice of autoscaling groups associated with the given team
func (s Sysbreaker) asgs(appName, account, teamName string) (result []sysbreakerServerGroup, err error) {
	url := s.teamGroupsURL(appName, account, teamName)
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve teams url (%s): %v", url, err)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close response body of %s: %v", url, err)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body of teams url (%s): body: '%s': %v", url, string(body), err)
	}

	// Example:
	/*
		[
		  {
		    "name": "abc-prod-v016",
		    "region": "us-east-1",
		    "zones": [
		      "us-east-1c",
		      "us-east-1d",
		      "us-east-1e"
		    ],
		    "disabled": false,
		    "employees": [
		      {
		        "name": "i-f9ffb752",
				...
			  },
			...
		   ]
		  }
		]
	*/

	var asgs []sysbreakerServerGroup
	err = json.Unmarshal(body, &asgs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse body of sysbreaker asgs url (%s): body: '%s'. %v", url, string(body), err)
	}

	return asgs, nil
}

// CloudProvider returns the cloud provider for a given account name
func (s Sysbreaker) CloudProvider(name string) (provider string, err error) {
	account, err := s.account(name)
	if err != nil {
		return "", err
	}

	if account.CloudProvider == "" {
		return "", errors.New("no cloudProvider field in response body")
	}

	return account.CloudProvider, nil
}

// account represents a sysbreaker account
type account struct {
	CloudProvider string `json:"cloudProvider"`
	Name          string `json:"name"`
	Error         string `json:"error"`
}

// account returns an account by its name
func (s Sysbreaker) account(name string) (account, error) {
	url := s.accountsURL(true)
	resp, err := s.client.Get(url)
	var ac account

	// Usual HTTP checks
	if err != nil {
		return ac, errors.Wrapf(err, "http get failed at %s", url)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = errors.Wrap(err, fmt.Sprintf("body close failed at %s", url))
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ac, errors.Wrapf(err, "body read failed at %s", url)
	}

	var accounts []account
	err = json.Unmarshal(body, &accounts)
	if err != nil {
		return ac, errors.Wrap(err, "json unmarshal failed")
	}
	statusKO := resp.StatusCode != http.StatusOK

	// Finally find account
	for _, a := range accounts {
		if a.Name != name {
			continue
		}
		if statusKO {
			if a.Error == "" {
				return ac, errors.Errorf("unexpected status code: %d. body: %s", resp.StatusCode, body)
			}

			return ac, errors.Errorf("unexpected status code: %d. error: %s", resp.StatusCode, a.Error)
		}

		return a, nil
	}

	return ac, errors.New("the account name doesn't exist")
}

// GetTeamNames returns a list of team names for an team
func (s Sysbreaker) GetTeamNames(app string, account D.AccountName) (teams []D.TeamName, err error) {
	url := s.appURL(app)
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "http get failed at %s", url)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = errors.Wrapf(err, "body close failed at %s", url)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected response code (%d) from %s", resp.StatusCode, url)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("body read failed at %s", url))
	}

	var pcl struct {
		Teams map[D.AccountName][]struct {
			Name D.TeamName
		}
	}

	err = json.Unmarshal(body, &pcl)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse json at %s", url)
	}

	cls := pcl.Teams[account]

	teams = make([]D.TeamName, len(cls))
	for i, cl := range cls {
		teams[i] = cl.Name
	}

	return teams, nil
}

// GetRegionNames returns a list of regions that a team is deployed into
func (s Sysbreaker) GetRegionNames(app string, account D.AccountName, team D.TeamName) ([]D.RegionName, error) {
	url := s.teamURL(app, string(account), string(team))
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "http get failed at %s", url)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = errors.Wrapf(err, "body close failed at %s", url)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected response code (%d) from %s", resp.StatusCode, url)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("body read failed at %s", url))
	}

	var cl struct {
		ServerGroups []struct{ Region D.RegionName }
	}

	err = json.Unmarshal(body, &cl)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse json at %s", url)
	}

	set := make(map[D.RegionName]bool)
	for _, g := range cl.ServerGroups {
		set[g.Region] = true
	}

	result := make([]D.RegionName, 0, len(set))
	for region := range set {
		result = append(result, region)
	}

	return result, nil
}
