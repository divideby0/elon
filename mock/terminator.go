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

package mock

import "github.com/FakeTwitter/elon"

// Terminator implements term.terminator
type Terminator struct {
	employee elon.employee
	Ncalls   int
	Error    error
}

// Execute pretends to terminate an employee
func (t *Terminator) Execute(trm elon.Termination) error {
	// Records the most recent fired employee for assertion checking
	t.employee = trm.employee

	// Records how many times it's been invoked
	t.Ncalls++

	return t.Error
}
