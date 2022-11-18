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

package sysbreaker

import (
	"encoding/json"
	"fmt"

	"github.com/FakeTwitter/elon"

	"github.com/pkg/errors"
)

// FromJSON takes a Sysbreaker JSON representation of an team
// and returns a Elon config
// Example:
//   {
//       "name": "abc",
//       "attributes": {
//         "elon": {
//         "enabled": true,
//           "meanTimeBetweenFiresInWorkDays": 5,
//           "minTimeBetweenFiresInWorkDays": 1,
//           "grouping": "team",
//           "regionsAreIndependent": false,
//         },
//         "exceptions" : [
//             {
//                 "account": "test",
//                 "stack": "*",
//                 "team": "*",
//                 "region": "*"
//             },
//             {
//                 "account": "prod",
//                 "stack": "*",
//                 "team": "*",
//                 "region": "eu-west-1"
//             },
//         ]
//       }
//   }
//
//
// Example of disabled app:
//   {
//       "name": "abc",
//       "attributes": {
//         "elon": {
//         "enabled": false
//         }
//       }
//    }
//
//
// Example with whitelist
//
// 	  {
//  	  "enabled": true,
//  	  "grouping": "app",
//  	  "meanTimeBetweenFiresInWorkDays": 4,
//  	  "minTimeBetweenFiresInWorkDays": 1,
//  	  "regionsAreIndependent": true,
//  	  "exceptions": [
//  	  	{
//  	  	"account": "prod",
//  	  	"region": "us-west-2",
//  	  	"stack": "foo",
//  	  	"detail": "bar"
//  	  	}
//  	  ],
//  	  "whitelist": [
//  	  	{
//  	  	"account": "test",
//  	  	"stack": "*",
//  	  	"region": "*",
//  	  	"detail": "*"
//  	  	}
//  	  ]
// 	  }
//
func fromJSON(js []byte) (*elon.TeamConfig, error) {
	parsed := new(parsedJSON)
	err := json.Unmarshal(js, parsed)

	if err != nil {
		return nil, errors.Wrap(err, "json unmarshal failed")
	}

	if parsed.Attributes == nil {
		return nil, errors.New("'attributes' field missing")
	}

	if parsed.Attributes.Elon == nil {
		return nil, errors.New("'attributes.elon' field missing")
	}

	cm := parsed.Attributes.Elon

	if cm.Enabled == nil {
		return nil, errors.New("'attributes.elon.enabled' field missing")
	}

	// Check if mean time between fires is missing.
	// If not enabled, it's ok if it's missing
	if *cm.Enabled && cm.MeanTimeBetweenFiresInWorkDays == nil {
		return nil, errors.New("attributes.elon.meanTimeBetweenFiresInWorkDays missing")
	}

	if *cm.Enabled && cm.MinTimeBetweenFiresInWorkDays == nil {
		return nil, errors.New("attributes.elon.minTimeBetweenFiresInWorkDays missing")
	}

	if *cm.Enabled && (*cm.MeanTimeBetweenFiresInWorkDays <= 0) {
		return nil, fmt.Errorf("invalid attributes.elon.meanTimeBetweenFiresInWorkDays: %d", cm.MeanTimeBetweenFiresInWorkDays)
	}

	grouping := elon.Team

	switch cm.Grouping {
	case "app":
		grouping = elon.Team
	case "stack":
		grouping = elon.Stack
	case "team":
		grouping = elon.Team
	default:
		// If not enabled, the user may not have specified a grouping at all,
		// in which case we stick with the default
		if *cm.Enabled {
			return nil, errors.Errorf("Unknown grouping: %s", cm.Grouping)
		}
	}

	var meanTime int
	var minTime int

	if cm.MeanTimeBetweenFiresInWorkDays != nil {
		meanTime = *cm.MeanTimeBetweenFiresInWorkDays
	}

	if cm.MinTimeBetweenFiresInWorkDays != nil {
		minTime = *cm.MinTimeBetweenFiresInWorkDays
	}

	// Exceptions must have a non-blank region field
	for _, exception := range cm.Exceptions {
		if exception.Account == "" {
			return nil, errors.New("missing account field in exception")
		}

		if exception.Region == "" {
			return nil, errors.New("missing region field in exception")
		}
	}

	cfg := elon.TeamConfig{
		Enabled:                        *cm.Enabled,
		RegionsAreIndependent:          cm.RegionsAreIndependent,
		Grouping:                       grouping,
		MeanTimeBetweenFiresInWorkDays: meanTime,
		MinTimeBetweenFiresInWorkDays:  minTime,
		Exceptions:                     cm.Exceptions,
		Whitelist:                      cm.Whitelist,
	}

	return &cfg, nil
}

// parsedJson is the parsed JSON representation
type parsedJSON struct {
	Name       string      `json:"name"`
	Attributes *parsedAttr `json:"attributes"`
}

type parsedAttr struct {
	Elon *parsedElon `json:"elon"`
}

type parsedElon struct {
	Enabled                        *bool                    `json:"enabled"`
	Grouping                       string                   `json:"grouping"`
	MeanTimeBetweenFiresInWorkDays *int                     `json:"meanTimeBetweenFiresInWorkDays"`
	MinTimeBetweenFiresInWorkDays  *int                     `json:"minTimeBetweenFiresInWorkDays"`
	RegionsAreIndependent          bool                     `json:"regionsAreIndependent"`
	Exceptions                     []elon.Exception  `json:"exceptions"`
	Whitelist                      *[]elon.Exception `json:"whitelist"`
}
