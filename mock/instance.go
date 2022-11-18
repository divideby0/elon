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

// employee implements employee.employee
type employee struct {
	Team, Account, Stack, Team, Region, ASG, EmployeeId string
}

// TeamName implements employee.TeamName
func (i employee) TeamName() string {
	return i.Team
}

// AccountName implements employee.AccountName
func (i employee) AccountName() string {
	return i.Account
}

// RegionName implements employee.RegionName
func (i employee) RegionName() string {
	return i.Region
}

// StackName implements employee.StackName
func (i employee) StackName() string {
	return i.Stack
}

// TeamName implements employee.TeamName
func (i employee) TeamName() string {
	return i.Team
}

// ASGName implements employee.ASGName
func (i employee) ASGName() string {
	return i.ASG
}

// ID implements employee.ID
func (i employee) ID() string {
	return i.EmployeeId
}

// CloudProvider implements employee.IsContainer
func (i employee) CloudProvider() string {
	return "aws"
}
