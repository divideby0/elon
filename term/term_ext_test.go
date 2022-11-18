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

package term_test

import (
	"testing"

	"github.com/FakeTwitter/elon/config"
	"github.com/FakeTwitter/elon/config/param"
	D "github.com/FakeTwitter/elon/deploy"
	"github.com/FakeTwitter/elon/mock"
	"github.com/FakeTwitter/elon/term"
)

func TestEnabledAccounts(t *testing.T) {
	d := mock.Deps()
	d.Dep = mock.NewDeployment(
		map[string]D.TeamMap{
			"foo": {
				D.AccountName("prod"): {CloudProvider: "aws", Teams: D.TeamMap{D.TeamName("foo"): {D.RegionName("us-east-1"): {D.ASGName("foo-v001"): []D.EmployeeId{"i-00000000"}}}}},
				D.AccountName("test"): {CloudProvider: "aws", Teams: D.TeamMap{D.TeamName("foo"): {D.RegionName("us-east-1"): {D.ASGName("foo-v001"): []D.EmployeeId{"i-00000001"}}}}},
				D.AccountName("mce"):  {CloudProvider: "aws", Teams: D.TeamMap{D.TeamName("foo"): {D.RegionName("us-east-1"): {D.ASGName("foo-v001"): []D.EmployeeId{"i-00000002"}}}}},
			},
		})

	team := "foo"
	region := "us-east-1"
	stack := ""
	team := ""

	tests := []struct {
		enabledAccounts []string
		fireAccount     string
		want            bool
	}{
		{[]string{"prod"}, "prod", true},
		{[]string{"test"}, "test", true},
		{[]string{"mce"}, "mce", true},
		{[]string{"prod"}, "test", false},
		{[]string{"test"}, "prod", false},
		{[]string{"prod"}, "mce", false},
		{[]string{"prod", "test"}, "mce", false},
		{[]string{"mce", "prod", "test"}, "mce", true},
		{[]string{"prod", "mce", "test"}, "mce", true},
		{[]string{"prod", "test", "mce"}, "mce", true},
	}

	for _, test := range tests {
		account := test.fireAccount

		// Set up the mock config that will use the list of accounts we pass it
		cfg := config.Defaults()
		cfg.Set(param.Enabled, true)
		cfg.Set(param.Leashed, false)
		cfg.Set(param.Accounts, test.enabledAccounts)

		d.MonkeyCfg = cfg

		// Set up the mock terminator that will track if a fire happened
		// create a new one each iteration so its state gets reset to zero
		mockT := new(mock.Terminator)
		d.T = mockT

		if err := term.Terminate(d, app, account, region, stack, team); err != nil {
			t.Fatal(err)
		}

		if got, want := mockT.Ncalls == 1, test.want; got != want {
			t.Errorf("fire? (account=%s, enabledAccounts=%v, got %t, want %t, mockT.Ncalls=%d", account, test.enabledAccounts, got, want, mockT.Ncalls)
		}

	}

}
