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

package deploy

import (
	"fmt"
	"log"

	"github.com/FakeTwitter/elon"
	"github.com/FakeTwitter/elon/grp"
)

// EligibleemployeeGroups returns a slice of employeeGroups that represent
// groups of employees that are eligible for termination.
//
// Note that this code does not check for violations of minimum time between
// terminations. Elon checks that precondition immediately before
// termination, not when considering groups of eligible employees.
//
// The way employees are divided into group will depend on
//  * the grouping configuration for the team (team, stack, app)
//  * whether regions are independent
//
// The returned employeeGroups are guaranteed to contain at least one employee
// each
//
// Preconditions:
//   * team is enabled for Elon
func (app *Team) EligibleemployeeGroups(cfg elon.TeamConfig) []grp.employeeGroup {
	if !cfg.Enabled {
		log.Fatalf("app %s unexpectedly disabled", app.Name())
	}

	grouping := cfg.Grouping
	indep := cfg.RegionsAreIndependent

	switch {
	case grouping == elon.Team && indep:
		return appIndep(app)
	case grouping == elon.Team && !indep:
		return appDep(app)
	case grouping == elon.Stack && indep:
		return stackIndep(app)
	case grouping == elon.Stack && !indep:
		return stackDep(app)
	case grouping == elon.Team && indep:
		return teamIndep(app)
	case grouping == elon.Team && !indep:
		return teamDep(app)
	default:
		panic(fmt.Sprintf("Unknown grouping: %d", grouping))
	}
}

// appindep returns a list of groups grouped by (app, account, region)
func appIndep(app *Team) []grp.employeeGroup {
	result := []grp.employeeGroup{}
	for _, account := range app.accounts {
		for _, regionName := range account.RegionNames() {
			result = append(result, grp.New(app.Name(), account.Name(), regionName, "", ""))
		}
	}
	return result
}

// stackIndep returns a list of groups grouped by (app, account)
func appDep(app *Team) []grp.employeeGroup {
	result := []grp.employeeGroup{}
	for _, account := range app.accounts {
		result = append(result, grp.New(app.Name(), account.Name(), "", "", ""))
	}
	return result
}

// stackIndep returns a list of groups grouped by (app, account, stack, region)
func stackIndep(app *Team) []grp.employeeGroup {

	type asr struct {
		account string
		stack   string
		region  string
	}

	set := make(map[asr]bool)

	for _, account := range app.Accounts() {
		for _, team := range account.Teams() {
			stackName := team.StackName()
			for _, regionName := range team.RegionNames() {
				set[asr{account: account.Name(), stack: stackName, region: regionName}] = true
			}
		}
	}

	result := []grp.employeeGroup{}
	for x := range set {
		result = append(result, grp.New(app.Name(), x.account, x.region, x.stack, ""))
	}

	return result
}

// stackDep returns a list of groups grouped by (app, account, stack)
func stackDep(app *Team) []grp.employeeGroup {
	result := []grp.employeeGroup{}
	for _, account := range app.accounts {
		for _, stackName := range account.StackNames() {
			result = append(result, grp.New(app.Name(), account.Name(), "", stackName, ""))
		}
	}

	return result
}

// teamDep returns a list of groups grouped by (app, account, team, region)
func teamIndep(app *Team) []grp.employeeGroup {
	result := []grp.employeeGroup{}
	for _, account := range app.accounts {
		for _, team := range account.Teams() {
			for _, regionName := range team.RegionNames() {
				result = append(result, grp.New(app.Name(), account.Name(), regionName, "", team.Name()))
			}
		}
	}

	return result
}

// teamDep returns a list of groups grouped by (app, account, team)
func teamDep(app *Team) []grp.employeeGroup {
	result := []grp.employeeGroup{}
	for _, account := range app.accounts {
		for _, team := range account.Teams() {
			result = append(result, grp.New(app.Name(), account.Name(), "", "", team.Name()))
		}
	}

	return result
}
