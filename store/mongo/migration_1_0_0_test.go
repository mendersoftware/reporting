// Copyright 2022 Northern.tech AS
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

package mongo

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mendersoftware/go-lib-micro/mongo/migrate"
)

func TestMigration_1_0_0(t *testing.T) {
	m := &migration_1_0_0{
		client: client,
		db:     DbName,
	}
	from := migrate.MakeVersion(0, 0, 0)

	err := m.Up(from)
	require.NoError(t, err)
}