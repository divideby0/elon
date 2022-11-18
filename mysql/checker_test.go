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

// +build docker

// The tests in this package use docker to test against a mysql:5.6 database
// By default, the tests are off unless you pass the "-tags docker" flag
// when running the test.

package mysql_test

import (
	"testing"
	"time"

	c "github.com/FakeTwitter/elon"
	"github.com/FakeTwitter/elon/mock"
	"github.com/FakeTwitter/elon/mysql"
)

var endHour = 15 // 3PM

// testSetup returns some values useful for test setup
func testSetup(t *testing.T) (ins c.employee, loc *time.Location, appCfg c.TeamConfig) {

	ins = mock.employee{
		Team:        "myapp",
		Account:    "prod",
		Stack:      "mystack",
		Team:    "myteam",
		Region:     "us-east-1",
		ASG:        "myapp-mystack-myteam-V123",
		EmployeeId: "i-a96a0166",
	}

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf(err.Error())
	}

	appCfg = c.TeamConfig{
		Enabled:                        true,
		RegionsAreIndependent:          true,
		MeanTimeBetweenFiresInWorkDays: 5,
		MinTimeBetweenFiresInWorkDays:  1,
		Grouping:                       c.Team,
		Exceptions:                     nil,
	}

	return

}

// TestCheckPermitted verifies check succeeds when no previous terminations in database
func TestCheckPermitted(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	m, err := mysql.New("localhost", port, "root", password, "elon")
	if err != nil {
		t.Fatal(err)
	}

	ins, loc, appCfg := testSetup(t)

	trm := c.Termination{employee: ins, Time: time.Now(), Leashed: false}

	err = m.Check(trm, appCfg, endHour, loc)

	if err != nil {
		t.Fatal(err)
	}
}

// TestCheckPermitted verifies check fails if commit is too recent
func TestCheckForbidden(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	m, err := mysql.New("localhost", port, "root", password, "elon")
	if err != nil {
		t.Fatal(err)
	}

	ins, loc, appCfg := testSetup(t)

	trm := c.Termination{employee: ins, Time: time.Now(), Leashed: false}

	// First check should succeed
	err = m.Check(trm, appCfg, endHour, loc)

	if err != nil {
		t.Fatal(err)
	}

	// Second check should fail
	err = m.Check(trm, appCfg, endHour, loc)
	if err == nil {
		t.Fatal("Check() succeeded when it should have failed")
	}

	if _, ok := err.(c.ErrViolatesMinTime); !ok {
		t.Fatalf("Expected Err.ViolatesMinTime, got %v", err)
	}
}

// When we are going to commit an unleashed termination, we only care
// about unleashed previous terminations
func TestCheckLeashed(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	m, err := mysql.New("localhost", port, "root", password, "elon")
	if err != nil {
		t.Fatal(err)
	}

	ins, loc, appCfg := testSetup(t)

	trm := c.Termination{employee: ins, Time: time.Now(), Leashed: true}

	// First check should succeed
	err = m.Check(trm, appCfg, endHour, loc)

	if err != nil {
		t.Fatal(err)
	}

	trm = c.Termination{employee: ins, Time: time.Now(), Leashed: false}

	// Second check should fail
	err = m.Check(trm, appCfg, endHour, loc)

	if err != nil {
		t.Fatalf("Should have allowed an unleashed termination after leashed: %v", err)
	}
}

// Check that only termination is permitted on concurrent attempts
func TestConcurrentChecks(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	m, err := mysql.New("localhost", port, "root", password, "elon")
	if err != nil {
		t.Fatal(err)
	}

	ins, loc, appCfg := testSetup(t)

	trm := c.Termination{employee: ins, Time: time.Now()}

	// Try to check twice. At least one should return an error
	ch := make(chan error, 2)

	go func() {
		// We use the "MySQL.CheckWithDelay" method which adds a delay between reading
		// from the database and writing to it, to increase the likelihood that
		// the two requests overlap
		ch <- m.CheckWithDelay(trm, appCfg, endHour, loc, 1*time.Second)
	}()

	go func() {
		ch <- m.Check(trm, appCfg, endHour, loc)
	}()

	var success int
	var txDeadlock int
	var violatesMinTime int
	for i := 0; i < 2; i++ {
		err := <-ch
		switch {
		case err == nil:
			success++
		case mysql.TxDeadlock(err):
			txDeadlock++
		case mysql.ViolatesMinTime(err):
			violatesMinTime++
		default:
			t.Fatalf("Unexpected error: %+v", err)
		}
	}

	if got, want := success, 1; got != want {
		t.Errorf("got %d succeses, want: %d", got, want)
	}
}

func TestCombinations(t *testing.T) {

	// Reference employee
	ins := mock.employee{
		Team:        "myapp",
		Account:    "prod",
		Stack:      "mystack",
		Team:    "myteam",
		Region:     "us-east-1",
		ASG:        "myapp-mystack-myteam-V123",
		EmployeeId: "i-a96a0166",
	}

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf(err.Error())
	}

	tests := []struct {
		desc    string
		grp     c.Group
		reg     bool // regions are independent
		ins     c.employee
		allowed bool // true if we can fire this employee after previous
	}{
		{"same team, should fail", c.Team, true, mock.employee{Team: "myapp", Account: "prod", Stack: "mystack", Team: "myteam", Region: "us-east-1", ASG: "myapp-mystack-myteam-V123"}, false},

		{"different team, should succeed", c.Team, true, mock.employee{Team: "myapp", Account: "prod", Stack: "mystack", Team: "otherteam", Region: "us-east-1", ASG: "myapp-mystack-myteam-V123"}, true},

		{"same stack should fail", c.Stack, true, mock.employee{Team: "myapp", Account: "prod", Stack: "mystack", Team: "otherteam", Region: "us-east-1", ASG: "myapp-mystack-myteam-V123"}, false},

		{"different stack, should succeed", c.Stack, true, mock.employee{Team: "myapp", Account: "prod", Stack: "otherstack", Team: "otherteam", Region: "us-east-1", ASG: "myapp-otherstack-myteam-V123"}, true},

		{"same app, should fail", c.Team, true, mock.employee{Team: "myapp", Account: "prod", Stack: "mystack", Team: "otherteam", Region: "us-east-1", ASG: "myapp-mystack-myteam-V123"}, false},

		{"different region, should succeed", c.Team, true, mock.employee{Team: "myapp", Account: "prod", Stack: "mystack", Team: "myteam", Region: "us-west-2", ASG: "myapp-mystack-myteam-V123"}, true},

		{"different region where regions are not independent, should fail", c.Team, false, mock.employee{Team: "myapp", Account: "prod", Stack: "mystack", Team: "myteam", Region: "us-west-2", ASG: "myapp-mystack-myteam-V123"}, false},
	}

	for _, tt := range tests {

		err := initDB()
		if err != nil {
			t.Fatal(err)
		}

		m, err := mysql.New("localhost", port, "root", password, "elon")
		if err != nil {
			t.Fatal(err)
		}
		cfg := c.TeamConfig{
			Enabled:                        true,
			RegionsAreIndependent:          tt.reg,
			MeanTimeBetweenFiresInWorkDays: 1,
			MinTimeBetweenFiresInWorkDays:  1,
			Grouping:                       tt.grp,
		}

		err = m.Check(c.Termination{employee: ins, Time: time.Now()}, cfg, endHour, loc)

		if err != nil {
			t.Fatal(err)
		}

		term := c.Termination{employee: tt.ins, Time: time.Now()}

		err = m.Check(term, cfg, endHour, loc)
		if tt.allowed && err != nil {
			t.Errorf("%s: got m.Check(%#v, %#v) = %+v, expected nil", tt.desc, term, cfg, err)
		}

		if !tt.allowed && err == nil {
			t.Errorf("%s: get m.Check(%#v, %#v) = nil, expected error", tt.desc, term, cfg)
		}

	}
}

func TestCheckMinTimeEnforced(t *testing.T) {

	cfg := c.TeamConfig{
		Enabled:                        true,
		RegionsAreIndependent:          true,
		MeanTimeBetweenFiresInWorkDays: 5,
		MinTimeBetweenFiresInWorkDays:  2,
		Grouping:                       c.Team,
	}

	// The current fire time
	now := "Thu Dec 17 11:35:00 2015 -0800"

	// Since MinTimeBetweenFiresInWorkDays is 1 here, then the most recent
	// fire permitted is the day before at endHour
	endHour := 15

	// Tue Dec 15 15:00:00 2015 -0800

	// Any fires later than that time will not be permitted
	// Boundary value testing!

	// this is a magic date used by go for parsing strings
	refDate := "Mon Jan  2 15:04:05 2006 -0700"
	tnow, err := time.Parse(refDate, now)
	if err != nil {
		t.Fatal(err)
	}

	ins := mock.employee{
		Team:        "myapp",
		Account:    "prod",
		Stack:      "mystack",
		Team:    "myteam",
		Region:     "us-east-1",
		ASG:        "myapp-mystack-myteam-V123",
		EmployeeId: "i-a96a0166",
	}

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		last    string
		allowed bool
	}{
		{"Tue Dec 15 15:01:00 2015 -0800", false},
		{"Tue Dec 15 14:59:59 2015 -0800", true},
	}

	for _, tt := range tests {

		err := initDB()
		if err != nil {
			t.Fatal(err)
		}

		m, err := mysql.New("localhost", port, "root", password, "elon")
		if err != nil {
			t.Fatal(err)
		}

		//
		// Write the initial termination
		//

		last, err := time.Parse("Mon Jan  2 15:04:05 2006 -0700", tt.last)
		if err != nil {
			t.Fatal(err)
		}
		err = m.Check(c.Termination{employee: ins, Time: last}, cfg, endHour, loc)
		if err != nil {
			t.Fatalf("Failed to write the initial termination, should always succeed: %v", err)
		}

		//
		// Write today's termination
		//

		err = m.Check(c.Termination{employee: ins, Time: tnow}, cfg, endHour, loc)

		switch err.(type) {
		case nil:
			if !tt.allowed {
				t.Fatalf("%s termination should have been forbidden, was allowed", tt.last)
			}
		case c.ErrViolatesMinTime:
			if tt.allowed {
				t.Errorf("%s termination should have been allowed, got: %v", tt.last, err)
			}
		default:
			t.Errorf("%s termination returned unexpected err: %v", tt.last, err)
		}
	}
}
