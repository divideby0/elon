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
	"runtime"
	"testing"
)

func TestASGAndTeams(t *testing.T) {
	nameOf := func(f interface{}) string {
		return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	}

	type tcase struct {
		appName     string
		accountName string
		regionName  string
		teamName string
		asgName     string
		ids         []string
	}

	makeTeamASG := func(tc tcase) (*Team, *ASG) {
		var team Team
		var account Account
		var team Team

		cloudProvider := "aws"

		asg := NewASG(tc.asgName, tc.regionName, tc.ids, &team)
		team = Team{tc.teamName, []*ASG{asg}, &account}
		account = Account{tc.accountName, []*Team{&team}, &app, cloudProvider}
		team = Team{tc.appName, []*Account{&account}}

		return &team, asg
	}

	type at struct {
		f    func(*ASG) string
		want string
	}

	type ct struct {
		f    func(*Team) string
		want string
	}

	var tests = []struct {
		scenario string
		t        tcase
		a        []at
		c        []ct
	}{
		{
			"stack and detail",
			tcase{"foo", "test", "us-east-1", "foo-staging-bar", "foo-staging-bar-v031", []string{"i-ff075688", "i-d9165a77"}},
			[]at{
				{(*ASG).Name, "foo-staging-bar-v031"},
				{(*ASG).TeamName, "foo"},
				{(*ASG).AccountName, "test"},
				{(*ASG).RegionName, "us-east-1"},
				{(*ASG).TeamName, "foo-staging-bar"},
				{(*ASG).StackName, "staging"},
				{(*ASG).DetailName, "bar"},
			},
			[]ct{
				{(*Team).Name, "foo-staging-bar"},
				{(*Team).TeamName, "foo"},
				{(*Team).AccountName, "test"},
				{(*Team).StackName, "staging"},
			},
		},
		{
			"no detail",
			tcase{"chaosguineapig", "prod", "eu-west-1", "chaosguineapig-staging", "chaosguineapig-staging-v000", []string{"i-7f40bbf5", "i-7a61d6f2"}},
			[]at{
				{(*ASG).Name, "chaosguineapig-staging-v000"},
				{(*ASG).TeamName, "chaosguineapig"},
				{(*ASG).AccountName, "prod"},
				{(*ASG).RegionName, "eu-west-1"},
				{(*ASG).TeamName, "chaosguineapig-staging"},
				{(*ASG).StackName, "staging"},
				{(*ASG).DetailName, ""},
			},
			[]ct{
				{(*Team).Name, "chaosguineapig-staging"},
				{(*Team).TeamName, "chaosguineapig"},
				{(*Team).AccountName, "prod"},
				{(*Team).StackName, "staging"},
			},
		},
		{
			"no stack",
			tcase{"chaosguineapig", "test", "eu-west-1", "chaosguineapig", "chaosguineapig-v030", []string{"i-7f40bbf5", "i-7a61d6f2"}},
			[]at{
				{(*ASG).Name, "chaosguineapig-v030"},
				{(*ASG).TeamName, "chaosguineapig"},
				{(*ASG).AccountName, "test"},
				{(*ASG).RegionName, "eu-west-1"},
				{(*ASG).TeamName, "chaosguineapig"},
				{(*ASG).StackName, ""},
				{(*ASG).DetailName, ""},
			},
			[]ct{
				{(*Team).Name, "chaosguineapig"},
				{(*Team).TeamName, "chaosguineapig"},
				{(*Team).AccountName, "test"},
				{(*Team).StackName, ""},
			},
		},
		{
			// We hit one case where there was a team with a name like foo-bar-v2, where the
			// asg had the same name: foo-bar-v2. The ASG had no push number, and the
			// detail looks like a push number.
			"detail looks like push number",
			tcase{"foo", "prod", "us-west-2", "foo-bar-v2", "foo-bar-v2", []string{"i-c7a513fc", "i-e06cfef1"}},
			[]at{
				{(*ASG).Name, "foo-bar-v2"},
				{(*ASG).TeamName, "foo"},
				{(*ASG).AccountName, "prod"},
				{(*ASG).RegionName, "us-west-2"},
				{(*ASG).TeamName, "foo-bar-v2"},
				{(*ASG).StackName, "bar"},
				{(*ASG).DetailName, "v2"},
			},
			[]ct{
				{(*Team).Name, "foo-bar-v2"},
				{(*Team).TeamName, "foo"},
				{(*Team).AccountName, "prod"},
				{(*Team).StackName, "bar"},
			},
		},
	}

	for _, tt := range tests {
		team, asg := makeTeamASG(tt.t)

		// ASG tests
		for _, att := range tt.a {
			if got, want := att.f(asg), att.want; got != want {
				t.Errorf("scenario %s: got %s()=%s, want: %s", tt.scenario, nameOf(att.f), got, want)
			}
		}

		// team tests
		for _, ctt := range tt.c {
			if got, want := ctt.f(team), ctt.want; got != want {
				t.Errorf("scenario %s: got %s()=%s, want: %s", tt.scenario, nameOf(ctt.f), got, want)
			}
		}
	}
}
