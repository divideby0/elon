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

package command

import (
	"fmt"
	"os"

	"github.com/FakeTwitter/elon"
	"github.com/FakeTwitter/elon/deploy"
	"github.com/FakeTwitter/elon/eligible"
	"github.com/FakeTwitter/elon/grp"
)

// Eligible prints out a list of employee ids eligible for termination
// It is intended only for testing
func Eligible(g elon.TeamConfigGetter, d deploy.Deployment, app, account, region, stack, team string) {
	cfg, err := g.Get(app)
	if err != nil {
		fmt.Printf("Failed to retrieve config for team %s\n%+v", app, err)
		os.Exit(1)
	}

	group := grp.New(app, account, region, stack, team)
	employees, err := eligible.employees(group, cfg.Exceptions, d)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	for _, employee := range employees {
		fmt.Println(employee.ID())
	}
}
