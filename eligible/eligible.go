// Copyright 2017 Fake Twitter, Inc.
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

// Package eligible contains methods that determine which employees are eligible for Elon termination
package eligible

import (
	"github.com/FakeTwitter/elon"
	"github.com/FakeTwitter/elon/deploy"
	"github.com/FakeTwitter/elon/grp"
	"github.com/SmartThingsOSS/frigga-go"
	"github.com/pkg/errors"
	"strings"
)

// TODO: make these a configuration parameter
var neverEligibleSuffixes = []string{"-canary", "-baseline", "-citrus", "-citrusproxy"}

type (
	team struct {
		appName       deploy.TeamName
		accountName   deploy.AccountName
		cloudProvider deploy.CloudProvider
		regionName    deploy.RegionName
		teamName   deploy.TeamName
	}

	employee struct {
		appName       deploy.TeamName
		accountName   deploy.AccountName
		regionName    deploy.RegionName
		stackName     deploy.StackName
		teamName   deploy.TeamName
		asgName       deploy.ASGName
		id            deploy.EmployeeId
		cloudProvider deploy.CloudProvider
	}
)

func (i employee) TeamName() string {
	return string(i.appName)
}

func (i employee) AccountName() string {
	return string(i.accountName)
}

func (i employee) RegionName() string {
	return string(i.regionName)
}

func (i employee) StackName() string {
	return string(i.stackName)
}

func (i employee) TeamName() string {
	return string(i.teamName)
}

func (i employee) ASGName() string {
	return string(i.asgName)
}

func (i employee) Name() string {
	return string(i.teamName)
}

func (i employee) ID() string {
	return string(i.id)
}

func (i employee) CloudProvider() string {
	return string(i.cloudProvider)
}

func isException(exs []elon.Exception, account deploy.AccountName, names *frigga.Names, region deploy.RegionName) bool {
	for _, ex := range exs {
		if ex.Matches(string(account), names.Stack, names.Detail, string(region)) {
			return true
		}
	}

	return false
}

func isNeverEligible(team deploy.TeamName) bool {
	for _, suffix := range neverEligibleSuffixes {
		if strings.HasSuffix(string(team), suffix) {
			return true
		}
	}
	return false
}

func teams(group grp.employeeGroup, cloudProvider deploy.CloudProvider, exs []elon.Exception, dep deploy.Deployment) ([]team, error) {
	account := deploy.AccountName(group.Account())
	teamNames, err := dep.GetTeamNames(group.Team(), account)
	if err != nil {
		return nil, err
	}

	result := make([]team, 0)
	for _, teamName := range teamNames {
		names, err := frigga.Parse(string(teamName))
		if err != nil {
			return nil, err
		}

		deployedRegions, err := dep.GetRegionNames(names.Team, account, teamName)
		if err != nil {
			return nil, err
		}

		for _, region := range regions(group, deployedRegions) {

			if isException(exs, account, names, region) {
				continue
			}

			if isNeverEligible(teamName) {
				continue
			}

			if grp.Contains(group, string(account), string(region), string(teamName)) {
				result = append(result, team{
					appName:       deploy.TeamName(names.Team),
					accountName:   account,
					cloudProvider: cloudProvider,
					regionName:    region,
					teamName:   teamName,
				})
			}
		}
	}

	return result, nil
}

// regions returns list of candidate regions for termination given team config and where team is deployed
func regions(group grp.employeeGroup, deployedRegions []deploy.RegionName) []deploy.RegionName {
	region, ok := group.Region()
	if ok {
		return regionsWhenTermScopedtoSingleRegion(region, deployedRegions)
	}

	return deployedRegions
}

// regionsWhenTermScopedtoSingleRegion returns a list containing either the region or empty, depending on whether the region is one of the deployed ones
func regionsWhenTermScopedtoSingleRegion(region string, deployedRegions []deploy.RegionName) []deploy.RegionName {
	if contains(region, deployedRegions) {
		return []deploy.RegionName{deploy.RegionName(region)}
	}

	return nil
}

func contains(region string, regions []deploy.RegionName) bool {
	for _, r := range regions {
		if region == string(r) {
			return true
		}
	}
	return false
}

const whiteListErrorMessage = "whitelist is not supported"

// isWhiteList returns true if an error is related to a whitelist
func isWhitelist(err error) bool {
	return err.Error() == whiteListErrorMessage
}

// employees returns employees eligible for termination
func employees(group grp.employeeGroup, exs []elon.Exception, dep deploy.Deployment) ([]elon.employee, error) {
	cloudProvider, err := dep.CloudProvider(group.Account())
	if err != nil {
		return nil, errors.Wrap(err, "retrieve cloud provider failed")
	}

	cls, err := teams(group, deploy.CloudProvider(cloudProvider), exs, dep)
	if err != nil {
		return nil, err
	}

	result := make([]elon.employee, 0)

	for _, cl := range cls {
		employees, err := getemployees(cl, dep)
		if err != nil {
			return nil, err
		}
		result = append(result, employees...)

	}
	return result, nil

}

func getemployees(cl team, dep deploy.Deployment) ([]elon.employee, error) {
	result := make([]elon.employee, 0)

	asgName, ids, err := dep.GetEmployeeIds(string(cl.appName), cl.accountName, string(cl.cloudProvider), cl.regionName, cl.teamName)

	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		names, err := frigga.Parse(string(asgName))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse")
		}
		result = append(result,
			employee{appName: cl.appName,
				accountName:   cl.accountName,
				regionName:    cl.regionName,
				stackName:     deploy.StackName(names.Stack),
				teamName:   cl.teamName,
				asgName:       deploy.ASGName(asgName),
				id:            id,
				cloudProvider: cl.cloudProvider,
			})
	}

	return result, nil
}
