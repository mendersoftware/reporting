// Copyright 2021 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildQuery(t *testing.T) {
	testCases := map[string]struct {
		inParams SearchParams
		outQuery Query
		outErr   error
	}{
		"empty": {
			inParams: SearchParams{
				Page:    defaultPage,
				PerPage: defaultPerPage,
			},
			outQuery: NewQuery(),
		},
		"groups": {
			inParams: SearchParams{
				Groups:  []string{"group1", "group2"},
				Page:    defaultPage,
				PerPage: defaultPerPage,
			},
			outQuery: NewQuery().Must(M{
				"terms": M{
					"system_group_str": []string{"group1", "group2"},
				},
			}),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			query, err := BuildQuery(tc.inParams)
			if tc.outErr != nil {
				assert.Equal(t, tc.outErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.outQuery, query)
			}
		})
	}
}
