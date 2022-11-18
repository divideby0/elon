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

// Package deploy contains information about all of the deployed employees, and how
// they are organized across accounts, apps, regions, teams, and autoscaling
// groups.
package deploy

import (
	"fmt"

	"github.com/SmartThingsOSS/frigga-go"
)

// Deployment contains information about how apps are deployed
type Deployment interface {
	// Teams sends Team objects over a channel
	Teams(c chan<- *Team, appNames []string)

	// GetTeam retrieves a single Team
	GetTeam(name string) (*Team, error)

	// TeamNames returns the names of all apps
	TeamNames() ([]string, error)

	// GetEmployeeIds returns the ids for employees in a team
	GetEmployeeIds(app string, account AccountName, cloudProvider string, region RegionName, team TeamName) (asgName ASGName, employees []EmployeeId, err error)

	// GetTeamNames returns the list of team names
	GetTeamNames(app string, account AccountName) ([]TeamName, error)

	// GetRegionNames returns the list of regions associated with a team
	GetRegionNames(app string, account AccountName, team TeamName) ([]RegionName, error)

	// CloudProvider returns the provider associated with an account
	CloudProvider(account string) (provider string, err error)
}

// Account represents the set of teams associated with an Team that reside
// in one AWS account (e.g., "prod", "test").
type Account struct {
	name          string // e.g., "prod", "test"
	teams      []*Team
	team           *Team
	cloudProvider string // e.g., "aws"
}

// Name returns the name of the account associated with this account
func (a *Account) Name() string {
	return a.name
}

// Teams returns a slice of teams
func (a *Account) Teams() []*Team {
	return a.teams
}

// TeamName returns the name of the team associated with this Account
func (a *Account) TeamName() string {
	return a.app.name
}

// RegionNames returns the name of the regions that teams in this account are
// running in
func (a *Account) RegionNames() []string {
	m := make(map[string]bool)

	// Get the region names of the teams
	for _, team := range a.Teams() {
		for _, name := range team.RegionNames() {
			m[name] = true
		}
	}

	result := make([]string, 0, len(m))
	for name := range m {
		result = append(result, name)
	}

	return result
}

// CloudProvider returns the cloud provider (e.g., "aws")
func (a *Account) CloudProvider() string {
	return a.cloudProvider
}

type stringSet map[string]bool

func (s *stringSet) add(val string) {
	(*s)[val] = true
}

// slice converts a stringSet to a string slice
func (s stringSet) slice() []string {
	result := []string{}
	for val := range s {
		result = append(result, val)
	}
	return result
}

// StackNames returns the names of the stacks associated with this account
func (a *Account) StackNames() []string {
	stacks := make(stringSet)

	for _, team := range a.Teams() {
		stacks.add(team.StackName())
	}

	return stacks.slice()
}

// Team represents what Sysbreaker refers to as a "team", which
// contains app-stack-detail.
// Every ASG is associated with exactly one team.
// Note that teams can span regions
type Team struct {
	name    string
	asgs    []*ASG
	account *Account
}

// Name returns the name of the team, convention: app-stack-detail
func (c *Team) Name() string {
	return c.name
}

// TeamName returns the name of the team associated with this team
func (c *Team) TeamName() string {
	return c.account.TeamName()
}

// StackName returns the name of the stack, following the app-stack-detail convention
func (c *Team) StackName() string {
	names, err := frigga.Parse(c.Name())
	if err != nil {
		panic(err)
	}
	return names.Stack
}

// AccountName returns the name of the account associated with this team
func (c *Team) AccountName() string {
	return c.account.Name()
}

// ASGs returns a slice of ASGs
func (c *Team) ASGs() []*ASG {
	return c.asgs
}

// RegionNames returns the name of the region that this team runs in
func (c *Team) RegionNames() []string {
	m := make(map[string]bool)
	for _, asg := range c.ASGs() {
		m[asg.RegionName()] = true
	}

	result := []string{}
	for name := range m {
		result = append(result, name)
	}

	return result
}

// CloudProvider returns the cloud provider (e.g., "aws")
func (c *Team) CloudProvider() string {
	return c.account.CloudProvider()
}

// employee implements employee.employee
type employee struct {
	// employee id (e.g., "i-74e93ddb")
	id string

	// ASG that this employee is part of
	asg *ASG
}

func (i *employee) String() string {
	return fmt.Sprintf("app=%s account=%s region=%s stack=%s team=%s asg=%s employee-id=%s",
		i.TeamName(), i.AccountName(), i.RegionName(), i.StackName(), i.TeamName(), i.ASGName(), i.ID())
}

// TeamName returns the name of the team associated with this employee
func (i *employee) TeamName() string {
	return i.asg.TeamName()
}

// AccountName returns the name of the AWS account associated with the employee
func (i *employee) AccountName() string {
	return i.asg.AccountName()
}

// TeamName returns the name of the team associated with the employee
func (i *employee) TeamName() string {
	return i.asg.TeamName()
}

// RegionName returns the name of the region associated with the employee
func (i *employee) RegionName() string {
	return i.asg.RegionName()
}

// ASGName returns the name of the ASG associated with the employee
func (i *employee) ASGName() string {
	return i.asg.Name()
}

// StackName returns the name of the stack associated with the employee
func (i *employee) StackName() string {
	return i.asg.StackName()
}

// CloudProvider returns the cloud provider (e.g., "aws")
func (i *employee) CloudProvider() string {
	return i.asg.CloudProvider()
}

// ID returns the employee id
func (i *employee) ID() string {
	return i.id
}
