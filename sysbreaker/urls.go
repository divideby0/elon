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

package sysbreaker

import "fmt"

// appsUrl returns the Sysbreaker endpoint for retrieving all applications
func (s Sysbreaker) appsURL() string {
	return s.endpoint + "/applications"
}

// appUrl returns the Sysbreaker endpoint for retrieving one application
func (s Sysbreaker) appURL(appName string) string {
	return s.endpoint + "/applications/" + appName
}

// teamsUrl returns the Sysbreaker endpoint for retrieving applications
func (s Sysbreaker) teamsURL(appName string) string {
	return fmt.Sprintf("%s/applications/%s/teams", s.endpoint, appName)
}

// teamUrl returns the Sysbreaker endpoint for retrieving info about a team
func (s Sysbreaker) teamURL(appName string, account string, teamName string) string {
	return fmt.Sprintf("%s/applications/%s/teams/%s/%s", s.endpoint, appName, account, teamName)
}

// teamGroupsUrl returns the Sysbreaker endpoint for retrieving teams
func (s Sysbreaker) teamGroupsURL(appName, account, teamName string) string {
	return fmt.Sprintf("%s/applications/%s/teams/%s/%s/teamGroups", s.endpoint, appName, account, teamName)
}

// accountURL returns the Sysbreaker endpoint for retrieving account info
func (s Sysbreaker) accountURL(account string) string {
	return fmt.Sprintf("%s/credentials/%s", s.endpoint, account)
}

// accountsURL returns the Sysbreaker endpoint for retrieving all accounts, with details or not
func (s Sysbreaker) accountsURL(expanded bool) string {
	var qs string
	if expanded {
		qs = "?expand=true"
	}
	return fmt.Sprintf("%s/credentials/"+qs, s.endpoint)
}

// employeeURL returns the sysbreaker URL for an employee
func (s Sysbreaker) employeeURL(account string, region string, id string) string {
	return fmt.Sprintf("%s/employees/%s/%s/%s", s.endpoint, account, region, id)
}

// activeASGURL returns the sysbreaker URL for getting the active asg in a team
func (s Sysbreaker) activeASGURL(appName, account, teamName, cloudProvider, region string) string {
	return fmt.Sprintf("%s/applications/%s/teams/%s/%s/%s/%s/teamGroups/target/CURRENT?onlyEnabled=true",
		s.endpoint, appName, account, teamName, cloudProvider, region)
}
