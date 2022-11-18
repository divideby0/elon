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

package mock

import (
	"github.com/pkg/errors"

	D "github.com/FakeTwitter/elon/deploy"
)

const cloudProvider = "aws"

// Dep returns a mock implementation of deploy.Deployment
// Dep has 4 apps: foo, bar, baz, quux
// Each team runs in 1 account:
//    foo, bar, baz run in prod
//    quux runs in test
// Each team has one team: foo-prod, bar-prod, baz-prod
// Each team runs in one region: us-east-1
// Each team contains 1 AZ with two employees
func Dep() D.Deployment {
	prod := D.AccountName("prod")
	test := D.AccountName("test")
	usEast1 := D.RegionName("us-east-1")

	return &Deployment{map[string]D.TeamMap{
		"foo":  {prod: D.AccountInfo{CloudProvider: cloudProvider, Teams: D.TeamMap{"foo-prod": {usEast1: {"foo-prod-v001": []D.EmployeeId{"i-d3e3d611", "i-63f52e25"}}}}}},
		"bar":  {prod: D.AccountInfo{CloudProvider: cloudProvider, Teams: D.TeamMap{"bar-prod": {usEast1: {"bar-prod-v011": []D.EmployeeId{"i-d7f06d45", "i-ce433cf1"}}}}}},
		"baz":  {prod: D.AccountInfo{CloudProvider: cloudProvider, Teams: D.TeamMap{"baz-prod": {usEast1: {"baz-prod-v004": []D.EmployeeId{"i-25b86646", "i-573d46d5"}}}}}},
		"quux": {test: D.AccountInfo{CloudProvider: cloudProvider, Teams: D.TeamMap{"quux-test": {usEast1: {"quux-test-v004": []D.EmployeeId{"i-25b866ab", "i-892d46d5"}}}}}},
	}}
}

// NewDeployment returns a mock implementation of deploy.Deployment
// Pass in a deploy.TeamMap, for example:
//  map[string]deploy.TeamMap{
// 		"foo":  deploy.TeamMap{"prod": {"foo-prod": {"us-east-1": {"foo-prod-v001": []string{"i-d3e3d611", "i-63f52e25"}}}}},
// 		"bar":  deploy.TeamMap{"prod": {"bar-prod": {"us-east-1": {"bar-prod-v011": []string{"i-d7f06d45", "i-ce433cf1"}}}}},
// 		"baz":  deploy.TeamMap{"prod": {"baz-prod": {"us-east-1": {"baz-prod-v004": []string{"i-25b86646", "i-573d46d5"}}}}},
// 		"quux": deploy.TeamMap{"test": {"quux-test": {"us-east-1": {"quux-test-v004": []string{"i-25b866ab", "i-892d46d5"}}}}},
// 	}
func NewDeployment(apps map[string]D.TeamMap) D.Deployment {
	return &Deployment{apps}
}

// Deployment implements deploy.Deployment interface
type Deployment struct {
	TeamMap map[string]D.TeamMap
}

// Teams implements deploy.Deployment.Teams
func (d Deployment) Teams(c chan<- *D.Team, apps []string) {
	defer close(c)

	for name, appmap := range d.TeamMap {
		c <- D.NewTeam(name, appmap)
	}
}

// GetTeamNames implements deploy.Deployment.GetTeamNames
func (d Deployment) GetTeamNames(app string, account D.AccountName) ([]D.TeamName, error) {
	result := make([]D.TeamName, 0)
	for team := range d.TeamMap[app][account].Teams {
		result = append(result, team)
	}

	return result, nil
}

// GetRegionNames implements deploy.Deployment.GetRegionNames
func (d Deployment) GetRegionNames(app string, account D.AccountName, team D.TeamName) ([]D.RegionName, error) {
	result := make([]D.RegionName, 0)
	for region := range d.TeamMap[app][account].Teams[team] {
		result = append(result, region)
	}

	return result, nil
}

// TeamNames implements deploy.Deployment.TeamNames
func (d Deployment) TeamNames() ([]string, error) {
	result := make([]string, len(d.TeamMap), len(d.TeamMap))
	i := 0
	for team := range d.TeamMap {
		result[i] = team
		i++
	}

	return result, nil
}

// GetTeam implements deploy.Deployment.GetTeam
func (d Deployment) GetTeam(name string) (*D.Team, error) {
	return D.NewTeam(name, d.TeamMap[name]), nil
}

// CloudProvider implements deploy.Deployment.CloudProvider
func (d Deployment) CloudProvider(account string) (string, error) {
	return cloudProvider, nil
}

// GetEmployeeIds implements deploy.Deployment.GetEmployeeIds
func (d Deployment) GetEmployeeIds(app string, account D.AccountName, cloudProvider string, region D.RegionName, team D.TeamName) (D.ASGName, []D.EmployeeId, error) {
	// Return an error if the team doesn't exist in the region

	appInfo, ok := d.TeamMap[app]
	if !ok {
		return "", nil, errors.Errorf("no team %s", app)
	}

	accountInfo, ok := appInfo[account]
	if !ok {
		return "", nil, errors.Errorf("app %s not deployed in account %s", app, account)
	}

	teamInfo, ok := accountInfo.Teams[team]
	if !ok {
		return "", nil, errors.Errorf("no team %s in app:%s, account:%s", team, app, account)
	}

	asgs, ok := teamInfo[region]
	if !ok {
		return "", nil, errors.Errorf("team %s in account %s not deployed in region %s", team, account, region)
	}

	employees := make([]D.EmployeeId, 0)

	// We assume there's only one asg, and retrieve the employees
	var asg D.ASGName
	var ids []D.EmployeeId

	for asg, ids = range asgs {
		for _, id := range ids {
			employees = append(employees, id)
		}
	}

	return asg, employees, nil
}
