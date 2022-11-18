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
	"reflect"
	"testing"

	"github.com/FakeTwitter/elon"
	"github.com/FakeTwitter/elon/grp"
)

type groupList []grp.employeeGroup

var grouptests = []struct {
	cfg    elon.TeamConfig
	groups []grp.employeeGroup
}{
	{conf(elon.Team, false), groupList{
		grp.New("mock", "prod", "", "", ""),
		grp.New("mock", "test", "", "", ""),
	}},
	{conf(elon.Team, true), groupList{
		grp.New("mock", "prod", "us-east-1", "", ""),
		grp.New("mock", "prod", "us-west-2", "", ""),
		grp.New("mock", "test", "us-east-1", "", ""),
		grp.New("mock", "test", "us-west-2", "", ""),
	}},
	{conf(elon.Stack, false), groupList{
		grp.New("mock", "prod", "", "prod", ""),
		grp.New("mock", "prod", "", "staging", ""),
		grp.New("mock", "test", "", "test", ""),
		grp.New("mock", "test", "", "beta", ""),
	}},
	{conf(elon.Stack, true), groupList{
		grp.New("mock", "prod", "us-east-1", "prod", ""),
		grp.New("mock", "prod", "us-west-2", "prod", ""),
		grp.New("mock", "prod", "us-east-1", "staging", ""),
		grp.New("mock", "prod", "us-west-2", "staging", ""),
		grp.New("mock", "test", "us-east-1", "test", ""),
		grp.New("mock", "test", "us-west-2", "test", ""),
		grp.New("mock", "test", "us-east-1", "beta", ""),
		grp.New("mock", "test", "us-west-2", "beta", ""),
	}},
	{conf(elon.Team, false), groupList{
		grp.New("mock", "prod", "", "", "mock-prod-a"),
		grp.New("mock", "prod", "", "", "mock-prod-b"),
		grp.New("mock", "prod", "", "", "mock-staging-a"),
		grp.New("mock", "prod", "", "", "mock-staging-b"),
		grp.New("mock", "test", "", "", "mock-test-a"),
		grp.New("mock", "test", "", "", "mock-test-b"),
		grp.New("mock", "test", "", "", "mock-beta-a"),
		grp.New("mock", "test", "", "", "mock-beta-b"),
	}},
	{conf(elon.Team, true), groupList{
		grp.New("mock", "prod", "us-east-1", "", "mock-prod-a"),
		grp.New("mock", "prod", "us-west-2", "", "mock-prod-a"),
		grp.New("mock", "prod", "us-east-1", "", "mock-prod-b"),
		grp.New("mock", "prod", "us-west-2", "", "mock-prod-b"),
		grp.New("mock", "prod", "us-east-1", "", "mock-staging-a"),
		grp.New("mock", "prod", "us-west-2", "", "mock-staging-a"),
		grp.New("mock", "prod", "us-east-1", "", "mock-staging-b"),
		grp.New("mock", "prod", "us-west-2", "", "mock-staging-b"),
		grp.New("mock", "test", "us-east-1", "", "mock-test-a"),
		grp.New("mock", "test", "us-west-2", "", "mock-test-a"),
		grp.New("mock", "test", "us-east-1", "", "mock-test-b"),
		grp.New("mock", "test", "us-west-2", "", "mock-test-b"),
		grp.New("mock", "test", "us-east-1", "", "mock-beta-a"),
		grp.New("mock", "test", "us-west-2", "", "mock-beta-a"),
		grp.New("mock", "test", "us-east-1", "", "mock-beta-b"),
		grp.New("mock", "test", "us-west-2", "", "mock-beta-b"),
	}},
}

func TestEligibleemployeeGroups(t *testing.T) {
	for i, tt := range grouptests {
		groups := mockTeam.EligibleemployeeGroups(tt.cfg)
		if len(tt.groups) != len(groups) {
			t.Errorf("test %d: incorrect number of groups. Expected: %d. Actual: %d", i, len(tt.groups), len(groups))
			continue
		}

		if !same(tt.groups, groups) {
			t.Errorf("test %d. Expected: %+v. Actual: %+v", i, tt.groups, groups)
		}
	}
}

//
// Test helper code
//

// conf creates a config file used for testing
func conf(grouping elon.Group, regionsAreIndependent bool) elon.TeamConfig {
	return elon.TeamConfig{
		Enabled:                        true,
		RegionsAreIndependent:          regionsAreIndependent,
		MeanTimeBetweenFiresInWorkDays: 5,
		MinTimeBetweenFiresInWorkDays:  1,
		Grouping:                       grouping,
	}
}

type groupSet map[grp.employeeGroup]bool

func (gs *groupSet) add(group grp.employeeGroup) {
	(*gs)[group] = true
}

func (gl groupList) toSet() groupSet {
	result := make(groupSet)
	for _, group := range gl {
		result.add(group)
	}
	return result
}

// same return true if the two lists of groups contain the same elements,
// independent of order
func same(x, y groupList) bool {
	sx := x.toSet()
	sy := y.toSet()
	return reflect.DeepEqual(sx, sy)
}

var usEast1 = RegionName("us-east-1")
var usWest2 = RegionName("us-west-2")

var mockTeam = NewTeam("mock", TeamMap{

	AccountName("prod"): {
		CloudProvider: "aws",
		Teams: TeamMap{
			TeamName("mock-prod-a"): {
				usEast1: {
					ASGName("mock-prod-a-v123"): []EmployeeId{"i-4a003cd0"},
				},
				usWest2: {
					ASGName("mock-prod-a-v111"): []EmployeeId{"i-efdc42dc"},
				},
			},
			TeamName("mock-prod-b"): {
				usEast1: {
					ASGName("mock-prod-b-v002"): []EmployeeId{"i-115ccc27"},
				},
				usWest2: {
					ASGName("mock-prod-b-v001"): []EmployeeId{"i-7881287e"},
				},
			},
			TeamName("mock-staging-a"): {
				usEast1: {
					ASGName("mock-staging-a-v123"): []EmployeeId{"i-ff8e7e4b"},
				},
				usWest2: {
					ASGName("mock-staging-a-v111"): []EmployeeId{"i-6eed18a4"},
				},
			},
			TeamName("mock-staging-b"): {
				usEast1: {
					ASGName("mock-staging-b-v002"): []EmployeeId{"i-13770e40"},
				},
				usWest2: {
					ASGName("mock-staging-b-v001"): []EmployeeId{"i-afb7595e"},
				},
			},
		},
	},
	AccountName("test"): {
		CloudProvider: "aws",
		Teams: TeamMap{
			TeamName("mock-test-a"): {
				usEast1: {
					ASGName("mock-test-a-v123"): []EmployeeId{"i-23b61f89"},
				},
				usWest2: {
					ASGName("mock-test-a-v111"): []EmployeeId{"i-fe7a0827"},
				},
			},
			TeamName("mock-test-b"): {
				usEast1: {
					ASGName("mock-test-b-v002"): []EmployeeId{"i-f581d5c3"},
				},
				usWest2: {
					ASGName("mock-test-b-v001"): []EmployeeId{"i-986e988a"},
				},
			},
			TeamName("mock-beta-a"): {
				usEast1: {
					ASGName("mock-beta-a-v123"): []EmployeeId{"i-4b359d5d"},
				},
				usWest2: {
					ASGName("mock-beta-a-v111"): []EmployeeId{"i-e751bdd2"},
				},
			},
			TeamName("mock-beta-b"): {
				usEast1: {
					ASGName("mock-beta-b-v002"): []EmployeeId{"i-e5eeba5e"},
				},
				usWest2: {
					ASGName("mock-beta-b-v001"): []EmployeeId{"i-76013ffb"},
				},
			},
		},
	},
})
