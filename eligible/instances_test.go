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

	"github.com/FakeTwitter/elon"
	D "github.com/FakeTwitter/elon/deploy"
	"github.com/FakeTwitter/elon/grp"
	"github.com/FakeTwitter/elon/mock"
)

// mockDeployment returns a deploy.Deployment object mock for testing
func mockDep() D.Deployment {
	usEast1 := D.RegionName("us-east-1")
	usWest2 := D.RegionName("us-west-2")
	return mock.NewDeployment(
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
						D.TeamName("mock-staging-a"): {
							usEast1: {
								D.ASGName("mock-staging-a-v123"): []D.EmployeeId{"i-ff8e7e4b"},
							},
							usWest2: {
								D.ASGName("mock-staging-a-v111"): []D.EmployeeId{"i-6eed18a4"},
							},
						},
						D.TeamName("mock-staging-b"): {
							usEast1: {
								D.ASGName("mock-staging-b-v002"): []D.EmployeeId{"i-13770e40"},
							},
							usWest2: {
								D.ASGName("mock-staging-b-v001"): []D.EmployeeId{"i-afb7595e"},
							},
						},
					},
				},
				D.AccountName("test"): {
					CloudProvider: "aws",
					Teams: D.TeamMap{
						D.TeamName("mock-test-a"): {
							usEast1: {
								D.ASGName("mock-test-a-v123"): []D.EmployeeId{"i-23b61f89"},
							},
							usWest2: {
								D.ASGName("mock-test-a-v111"): []D.EmployeeId{"i-fe7a0827"},
							},
						},
						D.TeamName("mock-test-b"): {
							usEast1: {
								D.ASGName("mock-test-b-v002"): []D.EmployeeId{"i-f581d5c3"},
							},
							usWest2: {
								D.ASGName("mock-test-b-v001"): []D.EmployeeId{"i-986e988a"},
							},
						},
						D.TeamName("mock-beta-a"): {
							usEast1: {
								D.ASGName("mock-beta-a-v123"): []D.EmployeeId{"i-4b359d5d"},
							},
							usWest2: {
								D.ASGName("mock-beta-a-v111"): []D.EmployeeId{"i-e751bdd2"},
							},
						},
						D.TeamName("mock-beta-b"): {
							usEast1: {
								D.ASGName("mock-beta-b-v002"): []D.EmployeeId{"i-e5eeba5e"},
							},
							usWest2: {
								D.ASGName("mock-beta-b-v001"): []D.EmployeeId{"i-76013ffb"},
							},
						},
					},
				},
			}})
}

func Testemployees(t *testing.T) {
	dep := mockDep()
	group := grp.New("mock", "prod", "us-east-1", "", "mock-prod-a")

	employees, err := employees(group, nil, dep)
	if err != nil {
		t.Fatal(err)
	}
	got, want := len(employees), 1
	if got != want {
		t.Fatalf("len(employees(group, nil, dep))=%v, want %v", got, want)
	}

	if employees[0].ID() != "i-4a003cd0" {
		t.Fatal("Expected id i-4a003cd0, got", employees[0].ID())
	}
}

func TestSimpleException(t *testing.T) {
	dep := mockDep()
	group := grp.New("mock", "prod", "us-east-1", "", "mock-prod-a")
	exs := []elon.Exception{{Account: "prod", Stack: "prod", Detail: "a", Region: "us-east-1"}}
	employees, err := employees(group, exs, dep)
	if err != nil {
		t.Fatal(err)
	}
	got, want := len(employees), 0
	if got != want {
		t.Fatalf("len(employees(group, exs, dep))=%v, want %v", got, want)
	}
}

func TestMultipleExceptions(t *testing.T) {
	team := abcloudMockDep()
	// Group across everything in prod
	group := grp.New("abcloud", "prod", "", "", "")
	exs := []elon.Exception{
		{Account: "prod", Stack: "batch", Detail: "", Region: "eu-west-1"},
		{Account: "prod", Stack: "ecom", Detail: "", Region: "us-west-2"},
		{Account: "prod", Stack: "", Detail: "", Region: "us-west-2"},
	}

	employees, err := employees(group, exs, app)
	if err != nil {
		t.Fatal(err)
	}
	got, want := len(employees), 6
	if got != want {
		t.Fatalf("len(employees(group, cfg, app))=%v, want %v", got, want)
	}

	// Ensure none of the excepted employees are in the list
	for _, employee := range employees {
		if employee.ID() == "i-8a1bd7ac" || employee.ID() == "i-2910a0e4" || employee.ID() == "i-b28a69c8" {
			t.Errorf("excepted employee is present: %v", employee)
		}
	}
}

// mockDep based on actual structure of abcloud
func abcloudMockDep() D.Deployment {
	usEast1 := D.RegionName("us-east-1")
	usWest2 := D.RegionName("us-west-2")
	euWest1 := D.RegionName("eu-west-1")
	return mock.NewDeployment(
		map[string]D.TeamMap{
			"abcloud": {
				D.AccountName("prod"): {
					CloudProvider: "aws",
					Teams: D.TeamMap{
						D.TeamName("abcloud"): {
							usEast1: {
								D.ASGName("abcloud-v123"): []D.EmployeeId{"i-7921a2f8"},
							},
							usWest2: {
								D.ASGName("abcloud-v123"): []D.EmployeeId{"i-8a1bd7ac"},
							},
							euWest1: {
								D.ASGName("abcloud-v123"): []D.EmployeeId{"i-87a90e92"},
							},
						},
						D.TeamName("abcloud-batch"): {
							usEast1: {
								D.ASGName("abcloud-batch-v123"): []D.EmployeeId{"i-2c25ab60"},
							},
							usWest2: {
								D.ASGName("abcloud-batch-v123"): []D.EmployeeId{"i-3bc40bdb"},
							},
							euWest1: {
								D.ASGName("abcloud-batch-v123"): []D.EmployeeId{"i-2910a0e4"},
							},
						},
						D.TeamName("abcloud-ecom"): {
							usEast1: {
								D.ASGName("abcloud-ecom-v123"): []D.EmployeeId{"i-ab9a4f10"},
							},
							usWest2: {
								D.ASGName("abcloud-ecom-v123"): []D.EmployeeId{"i-b28a69c8"},
							},
							euWest1: {
								D.ASGName("abcloud-ecom-v123"): []D.EmployeeId{"i-4fa09365"},
							},
						},
					},
				},
			},
		},
	)
}
