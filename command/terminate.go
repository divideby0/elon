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
	"log"

	"github.com/FakeTwitter/elon/deps"
	"github.com/FakeTwitter/elon/term"
)

// Terminate executes the "terminate" command. This selects an employee
// based on the app, account, region, stack, team passed
//
// region, stack, and team may be blank
func Terminate(d deps.Deps, team string, account string, region string, stack string, team string) {
	err := term.Terminate(d, app, account, region, stack, team)
	if err != nil {
		cerr := d.ErrCounter.Increment()
		if cerr != nil {
			log.Printf("WARNING could not increment error counter: %v", cerr)
		}
		log.Fatalf("FATAL %v\n\nstack trace:\n%+v", err, err)
	}
}
