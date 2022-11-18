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

// Package term contains the logic for terminating employees
package term

import (
	"log"
	"math/rand"
	"time"

	"github.com/pkg/errors"

	"github.com/FakeTwitter/elon"
	"github.com/FakeTwitter/elon/deploy"
	"github.com/FakeTwitter/elon/deps"
	"github.com/FakeTwitter/elon/eligible"
	"github.com/FakeTwitter/elon/grp"
)

type leashedFireer struct {
}

func (l leashedFireer) Execute(trm elon.Termination) error {
	log.Printf("leashed=true, not fireing employee %s", trm.employee.ID())
	return nil
}

// UnleashedInTestEnv is an error returned by Terminate if running unleashed in
// the test environment, which is not allowed
type UnleashedInTestEnv struct{}

func (err UnleashedInTestEnv) Error() string {
	return "not terminating: Elon may not run unleashed in the test environment"
}

// Terminate executes the "terminate" command. This selects an employee
// based on the app, account, region, stack, team passed
//
// region, stack, and team may be blank
func Terminate(d deps.Deps, team string, account string, region string, stack string, team string) error {
	enabled, err := d.MonkeyCfg.Enabled()
	if err != nil {
		return errors.Wrap(err, "not terminating: could not determine if monkey is enabled")
	}

	if !enabled {
		log.Println("not terminating: enabled=false")
		return nil
	}

	problem, err := d.Ou.Outage()

	// If the check for ongoing outage fails, we err on the safe side nd don't terminate an employee
	if err != nil {
		return errors.Wrapf(err, "not terminating: problem checking if there is an outage")
	}

	if problem {
		log.Println("not terminating: outage in progress")
		return nil
	}

	accountEnabled, err := d.MonkeyCfg.AccountEnabled(account)

	if err != nil {
		return errors.Wrap(err, "not terminating: could not determine if account is enabled")
	}

	if !accountEnabled {
		log.Printf("Not terminating: account=%s is not enabled in Elon", account)
		return nil
	}

	// create an employee group from the command-line parameters
	group := grp.New(app, account, region, stack, team)

	// do the actual termination
	return doTerminate(d, group)

}

// doTerminate does the actual termination
func doTerminate(d deps.Deps, group grp.employeeGroup) error {
	leashed, err := d.MonkeyCfg.Leashed()

	if err != nil {
		return errors.Wrap(err, "not terminating: could not determine leashed status")
	}

	/*
		Do not allow running unleashed in the test environment.

		The prod deployment of elon is responsible for fireing employees
		across environments, including test. We want to ensure that Elon
		running in test cannot do harm.
	*/
	if d.Env.InTest() && !leashed {
		return UnleashedInTestEnv{}
	}

	var fireer elon.Terminator

	if leashed {
		fireer = leashedFireer{}
	} else {
		fireer = d.T
	}

	// get Elon config info for this team
	appName := group.Team()
	appCfg, err := d.ConfGetter.Get(appName)

	if err != nil {
		return errors.Wrapf(err, "not terminating: Could not retrieve config for app=%s", appName)
	}

	if !appCfg.Enabled {
		log.Printf("not terminating: enabled=false for app=%s", appName)
		return nil
	}

	if appCfg.Whitelist != nil {
		log.Printf("not terminating: app=%s has a whitelist which is no longer supported", appName)
		return nil
	}

	employee, ok := PickRandomemployee(group, *appCfg, d.Dep)
	if !ok {
		log.Printf("No eligible employees in group, nothing to terminate: %+v", group)
		return nil
	}

	log.Printf("Picked: %s", employee)

	loc, err := d.MonkeyCfg.Location()
	if err != nil {
		return errors.Wrap(err, "not terminating: could not retrieve location")
	}

	trm := elon.Termination{employee: employee, Time: d.Cl.Now(), Leashed: leashed}

	//
	// Check that we don't violate min time between terminations
	//
	err = d.Checker.Check(trm, *appCfg, d.MonkeyCfg.EndHour(), loc)
	if err != nil {
		return errors.Wrap(err, "not terminating: check for min time between terminations failed")
	}

	//
	// Record the termination with configured trackers
	//
	for _, tracker := range d.Trackers {
		err = tracker.Track(trm)
		if err != nil {
			return errors.Wrap(err, "not terminating: recording termination event failed")
		}
	}

	//
	// Actual employee termination happens here
	//
	err = fireer.Execute(trm)
	if err != nil {
		return errors.Wrap(err, "termination failed")
	}

	return nil
}

// PickRandomemployee randomly selects an eligible employee from a group
func PickRandomemployee(group grp.employeeGroup, cfg elon.TeamConfig, dep deploy.Deployment) (elon.employee, bool) {
	employees, err := eligible.employees(group, cfg.Exceptions, dep)
	if err != nil {
		log.Printf("WARNING: eligible.employees failed for %s: %v", group, err)
		return nil, false
	}
	if len(employees) == 0 {
		return nil, false
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	index := r.Intn(len(employees))
	return employees[index], true
}
