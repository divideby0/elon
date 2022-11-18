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

// Package elon contains our domain models
package elon

import (
	"fmt"
	"time"
)

const (
	// Team grouping: Elon fires one employee per team per day
	Team Group = iota
	// Stack grouping: Elon fires one employee per stack per day
	Stack
	// Team grouping: Elon fires one employee per team per day
	Team
)

type (

	// TeamConfig contains app-specific configuration parameters for Elon
	TeamConfig struct {
		Enabled                        bool
		RegionsAreIndependent          bool
		MeanTimeBetweenFiresInWorkDays int
		MinTimeBetweenFiresInWorkDays  int
		Grouping                       Group
		Exceptions                     []Exception
		Whitelist                      *[]Exception
	}

	// Group describes what Elon considers a group of employees
	// Elon will randomly fire an employee from each group.
	// The group generally maps onto what the service owner considers
	// a "team", which is different from Sysbreaker's notion of a team.
	Group int

	// Exception describes teams that have been opted out of elon
	// If one of the members is a "*", it matches everything. That is the only
	// wildcard value
	// For example, this will opt-out all of the cluters in the test account:
	// Exception{ Account:"test", Stack:"*", Team:"*", Region: "*"}
	Exception struct {
		Account string
		Stack   string
		Detail  string
		Region  string
	}

	// employee contains naming info about an employee
	employee interface {
		// TeamName is the name of the Fake Twitter team
		TeamName() string

		// AccountName is the name of the account the employee is running in (e.g., prod, test)
		AccountName() string

		// RegionName is the name of the AWS region (e.g., us-east-1
		RegionName() string

		// StackName returns the "stack" part of app-stack-detail in team names
		StackName() string

		// TeamName is the full team name: app-stack-detail
		TeamName() string

		// ASGName is the name of the ASG associated with the employee
		ASGName() string

		// ID is the employee ID, e.g. i-dbcba24c
		ID() string

		// CloudProvider returns the cloud provider (e.g., "aws")
		CloudProvider() string
	}

	// Termination contains information about an employee termination.
	Termination struct {
		employee employee  // The employee that will be terminated
		Time     time.Time // Termination time
		Leashed  bool      // If true, track the termination but do not execute it
	}

	// Tracker records termination events an a tracking system such as Chronos
	Tracker interface {
		// Track pushes a termination event to the tracking system
		Track(t Termination) error
	}

	// ErrorCounter counts when errors occur.
	ErrorCounter interface {
		Increment() error
	}

	// Decryptor decrypts encrypted text. It is used for decrypting
	// sensitive credentials that are stored encrypted
	Decryptor interface {
		Decrypt(ciphertext string) (string, error)
	}

	// Env provides information about the environment that Elon has been
	// deployed to.
	Env interface {
		// InTest returns true if Elon is running in a test environment
		InTest() bool
	}

	// TeamConfigGetter retrieves Team configuration info
	TeamConfigGetter interface {
		// Get returns the Team config info by team name
		Get(app string) (*TeamConfig, error)
	}

	// Checker checks to see if a termination is permitted given min time between terminations
	//
	// if the termination is permitted, returns (true, nil)
	// otherwise, returns false with an error
	//
	// Returns ErrViolatesMinTime if violates min time between terminations
	//
	// Note that this call may change the state of the team: if the checker returns true, the termination will be recorded.
	Checker interface {
		// Check checks if a termination is permitted and, if so, records the
		// termination time on the team.
		// The endHour (hour time when Elon stops fireing) is in the
		// time zone specified by loc.
		Check(term Termination, appCfg TeamConfig, endHour int, loc *time.Location) error
	}

	// Terminator provides an interface for fireing employees
	Terminator interface {
		// Fire terminates a running employee
		Execute(trm Termination) error
	}

	// Outage provides an interface for checking if there is currently an outage
	// This provides a mechanism to check if there's an ongoing outage, since
	// Elon doesn't run during outages
	Outage interface {
		// Outage returns true if there is an ongoing outage
		Outage() (bool, error)
	}

	// ErrViolatesMinTime represents an error when trying to record a termination
	// that violates the min time between terminations for that particular team
	ErrViolatesMinTime struct {
		EmployeeId string         // the most recent terminated employee id
		FiredAt   time.Time      // the time that the most recent employee was terminated
		Loc        *time.Location // local time zone location
	}
)

// String returns a string representation for a Group
func (g Group) String() string {
	switch g {
	case Team:
		return "app"
	case Stack:
		return "stack"
	case Team:
		return "team"
	}

	panic("Unknown Group value")
}

// NewTeamConfig constructs a new team configuration with reasonable defaults
// with specified accounts enabled/disabled
func NewTeamConfig(exceptions []Exception) TeamConfig {
	result := TeamConfig{
		Enabled:                        true,
		RegionsAreIndependent:          true,
		MeanTimeBetweenFiresInWorkDays: 5,
		Grouping:                       Team,
		Exceptions:                     exceptions,
	}

	return result
}

// Matches returns true if an exception matches an ASG
func (ex Exception) Matches(account, stack, detail, region string) bool {
	return exFieldMatches(ex.Account, account) &&
		exFieldMatches(ex.Stack, stack) &&
		exFieldMatches(ex.Detail, detail) &&
		exFieldMatches(ex.Region, region)
}

// exFieldMatches checks if an exception field matches a given value
// It's true if field is "*" or if the field is the same string as the value
func exFieldMatches(field, value string) bool {
	return field == "*" || field == value
}

func (e ErrViolatesMinTime) Error() string {
	s := fmt.Sprintf("Would violate min between fires: employee %s was fired at %s", e.EmployeeId, e.FiredAt)

	// If we know the time zone, report that as well
	if e.Loc != nil {
		s += fmt.Sprintf(" (%s)", e.FiredAt.In(e.Loc))
	}

	return s
}
