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

package eligible

import (
	"testing"

	D "github.com/FakeTwitter/elon/deploy"
	"github.com/FakeTwitter/elon/grp"
	"github.com/FakeTwitter/elon/mock"
)

// Test that canaries are not considered eligible employees
func TestNoFireCanaries(t *testing.T) {
	usEast1 := D.RegionName("us-east-1")
	usWest2 := D.RegionName("us-west-2")

	dep := mock.NewDeployment(
		map[string]D.TeamMap{
			"mock": {
				D.AccountName("prod"): {
					CloudProvider: "aws",
					Teams: D.TeamMap{
						D.TeamName("mock-prod-a"): {
							usEast1: {
								D.ASGName("mock-prod-a-v123"): []D.EmployeeId{"i-4a003cd0"},
							},
							usWest2: {
								D.ASGName("mock-prod-a-v111"): []D.EmployeeId{"i-efdc42dc"},
							},
						},
						D.TeamName("mock-prod-b"): {
							usEast1: {
								D.ASGName("mock-prod-b-v002"): []D.EmployeeId{"i-115ccc27"},
							},
							usWest2: {
								D.ASGName("mock-prod-b-v001"): []D.EmployeeId{"i-7881287e"},
							},
						},
						D.TeamName("mock-prod-b-baseline"): {
							usEast1: {
								D.ASGName("mock-prod-b-baseline-v012"): []D.EmployeeId{"i-e71a94d0"},
							},
							usWest2: {
								D.ASGName("mock-prod-b-baseline-v011"): []D.EmployeeId{"i-69211000"},
							},
						},
						D.TeamName("mock-prod-b-canary"): {
							usEast1: {
								D.ASGName("mock-prod-b-canary-v012"): []D.EmployeeId{"i-18d2e1b6"},
							},
							usWest2: {
								D.ASGName("mock-prod-b-canary-v011"): []D.EmployeeId{"i-63bda865"},
							},
						},
						D.TeamName("mock-prod-a-citrus"): {
							usEast1: {
								D.ASGName("mock-prod-b-citrus-v014"): []D.EmployeeId{"i-d26e6af1"},
							},
							usWest2: {
								D.ASGName("mock-prod-b-citrus-v013"): []D.EmployeeId{"i-1db216c3"},
							},
						},
						D.TeamName("mock-prod-a-citrusproxy"): {
							usEast1: {
								D.ASGName("mock-prod-b-citrusproxy-v020"): []D.EmployeeId{"i-c57ad10c"},
							},
							usWest2: {
								D.ASGName("mock-prod-b-citrusproxy-v017"): []D.EmployeeId{"i-6fba090b"},
							},
						},
					},
				},
			},
		},
	)

	// Group is all employees in mock app, prod group
	group := grp.New("mock", "prod", "", "", "")
	employees, err := employees(group, nil, dep)
	if err != nil {
		t.Fatal(err)
	}

	got, want := len(employees), 4
	if got != want {
		t.Fatalf("len(Eligibleemployees(group, cfg, deployInfo))=%d, want %d", got, want)
	}
}
