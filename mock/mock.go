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

// Package mock contains helper functions for generating mock objects
// for testing
package mock

import D "github.com/FakeTwitter/elon/deploy"

// TeamFactory creates Team objects used for testing
type TeamFactory struct {
}

// Team creates a mock Team
func (factory TeamFactory) Team() *D.Team {

	var m = D.TeamMap{
		"prod": D.AccountInfo{
			CloudProvider: "aws",
			Teams: D.TeamMap{
				"abc-prod": {
					"us-east-1": {
						"abc-prod-v017": []D.EmployeeId{"i-f60b22e8", "i-1b17963b", "i-7c0c8af4"},
					},
					"us-west-2": {
						"abc-prod-v017": []D.EmployeeId{"i-8b42d04e", "i-52ead2f0", "i-b6261b80"},
					},
				},
			},
		},
		"test": D.AccountInfo{
			CloudProvider: "aws",
			Teams: D.TeamMap{
				"abc-beta": {
					"us-east-1": {
						"abc-beta-v031": []D.EmployeeId{"i-c8a5458c", "i-61f55db3", "i-6a820363"},
						"abc-beta-v030": []D.EmployeeId{"i-c41206b7", "i-c8a5458c", "i-6a820363"},
					},
				},
			},
		},
	}
	return D.NewTeam("abc", m)
}
