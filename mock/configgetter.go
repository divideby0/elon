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

// ConfigGetter implements elon.Getter
type ConfigGetter struct {
	Config elon.TeamConfig
}

// NewConfigGetter returns a mock config getter that always returns the specified config
func NewConfigGetter(config elon.TeamConfig) ConfigGetter {
	return ConfigGetter{Config: config}
}

// DefaultConfigGetter returns a mock config getter that always returns the same config
func DefaultConfigGetter() ConfigGetter {
	return ConfigGetter{
		Config: elon.TeamConfig{
			Enabled:                        true,
			RegionsAreIndependent:          true,
			MeanTimeBetweenFiresInWorkDays: 5,
			MinTimeBetweenFiresInWorkDays:  1,
			Grouping:                       elon.Team,
			Exceptions:                     nil,
		},
	}
}

// Get implements elon.Getter.Get
func (c ConfigGetter) Get(app string) (*elon.TeamConfig, error) {
	return &c.Config, nil
}
