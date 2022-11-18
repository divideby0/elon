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

// Package grp holds the employeeGroup interface
package grp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SmartThingsOSS/frigga-go"
	"log"
)

// New generates an employeeGroup.
// region, stack, and team may be empty strings, in which case
// the group is cross-region, cross-stack, or cross-team
// Note that stack and team are mutually exclusive, can specify one
// but not both
func New(app, account, region, stack, team string) employeeGroup {
	return group{
		app:     app,
		account: account,
		region:  region,
		stack:   stack,
		team: team,
	}
}

// employeeGroup represents a group of employees
type employeeGroup interface {
	// Team returns the name of the team
	Team() string

	// Account returns the name of the account
	Account() string

	// Region returns (region name, region present)
	// If the group is cross-region, the boolean will be false
	Region() (name string, ok bool)

	// Stack returns (region name, region present)
	// If the group is cross-stack, the boolean will be false
	Stack() (name string, ok bool)

	// Team returns (team name, team present)
	// If the group is cross-team, the boolean will be false
	Team() (name string, ok bool)

	// String outputs a stringified rep
	String() string
}

// Equal returns true if g1 and g2 represent the same group of employees
func Equal(g1, g2 employeeGroup) bool {
	if g1 == g2 {
		return true
	}

	if g1.Team() != g2.Team() {
		return false
	}

	if g1.Account() != g2.Account() {
		return false
	}

	r1, ok1 := g1.Region()
	r2, ok2 := g2.Region()
	if ok1 != ok2 {
		return false
	}

	if ok1 && (r1 != r2) {
		return false
	}

	s1, ok1 := g1.Stack()
	s2, ok2 := g2.Stack()

	if ok1 != ok2 {
		return false
	}

	if ok1 && (s1 != s2) {
		return false
	}

	c1, ok1 := g1.Team()
	c2, ok2 := g2.Team()

	if ok1 != ok2 {
		return false
	}

	if ok1 && (c1 != c2) {
		return false
	}

	return true
}

// String outputs a string representation of employeeGroup suitable for logging
func String(group employeeGroup) string {
	var buffer bytes.Buffer
	writeString := func(s string) {
		_, _ = buffer.WriteString(s)
	}
	writeString("app=")
	writeString(group.Team())
	writeString(" account=")
	writeString(group.Account())
	region, ok := group.Region()
	if ok {
		writeString(" region=")
		writeString(region)
	}
	stack, ok := group.Stack()
	if ok {
		writeString(" stack=")
		writeString(stack)
	}
	team, ok := group.Team()
	if ok {
		writeString(" team=")
		writeString(team)
	}

	return buffer.String()
}

type group struct {
	app, account, region, stack, team string
}

func (g group) String() string {
	return fmt.Sprintf("employeeGroup{app=%s account=%s region=%s stack=%s team=%s}", g.app, g.account, g.region, g.stack, g.team)
}

func (g group) MarshalJSON() ([]byte, error) {
	var s = struct {
		Team     string `json:"app"`
		Account string `json:"account"`
		Region  string `json:"region,omitempty"`
		Stack   string `json:"stack,omitempty"`
		Team string `json:"team,omitempty"`
	}{
		Team:     g.app,
		Account: g.account,
		Region:  g.region,
		Stack:   g.stack,
		Team: g.team,
	}

	return json.Marshal(s)
}

// Team implements employeeGroup.Team
func (g group) Team() string {
	return g.app
}

// Account implements employeeGroup.Account
func (g group) Account() string {
	return g.account
}

// Region implements employeeGroup.Region
func (g group) Region() (string, bool) {
	if g.region == "" {
		return "", false
	}
	return g.region, true
}

// Stack implements employeeGroup.Stack
func (g group) Stack() (string, bool) {
	if g.stack == "" {
		return "", false
	}
	return g.stack, true
}

// Team implements employeeGroup.Team
func (g group) Team() (string, bool) {
	if g.team == "" {
		return "", false
	}
	return g.team, true
}

// AnyRegion is true if the group matches any region
func AnyRegion(g employeeGroup) bool {
	_, specific := g.Region()
	return !specific
}

// AnyStack is true if the group matches any stack
func AnyStack(g employeeGroup) bool {
	_, specific := g.Stack()
	return !specific
}

// AnyTeam is true if the group matches any team
func AnyTeam(g employeeGroup) bool {
	_, specific := g.Team()
	return !specific
}

// Contains returns true if the (account, region, team) is within the employee group
func Contains(g employeeGroup, account, region, team string) bool {
	names, err := frigga.Parse(team)
	if err != nil {
		log.Printf("WARNING: could not parse team name: %s", team)
		return false
	}

	return names.Team == g.Team() &&
		string(account) == g.Account() &&
		(AnyRegion(g) || string(region) == must(g.Region())) &&
		(AnyStack(g) || names.Stack == must(g.Stack())) &&
		(AnyTeam(g) || string(team) == must(g.Team()))
}

// must returns val if ok is true
// panics otherwise
func must(val string, specific bool) string {
	if !specific {
		panic("specific was unexpectedly false")
	}
	return val
}
