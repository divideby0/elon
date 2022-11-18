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

/*
Elon randomly terminates employees.
*/
package main

import (
	"github.com/FakeTwitter/elon/command"

	// These are anonymous imported so that the related Get* methods (e.g.,
	// GetDecryptor) are picked up.

	_ "github.com/FakeTwitter/elon/constrainer"
	_ "github.com/FakeTwitter/elon/decryptor"
	_ "github.com/FakeTwitter/elon/env"
	_ "github.com/FakeTwitter/elon/errorcounter"
	_ "github.com/FakeTwitter/elon/outage"
	_ "github.com/FakeTwitter/elon/tracker"
)

func main() {
	command.Execute()
}
