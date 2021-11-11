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

package indexer

import (
	"context"

	"github.com/mendersoftware/reporting/client/deviceauth"
	"github.com/mendersoftware/reporting/client/inventory"
	"github.com/mendersoftware/reporting/client/nats"
	"github.com/mendersoftware/reporting/model"
	"github.com/mendersoftware/reporting/store"
)

//nolint:lll
//go:generate ../../x/mockgen.sh
type Indexer interface {
	GetJobs(ctx context.Context, jobs chan *model.Job) error
	ProcessJobs(ctx context.Context, jobs []*model.Job)
}

type indexer struct {
	store     store.Store
	nats      nats.Client
	devClient deviceauth.Client
	invClient inventory.Client
}

func NewIndexer(
	store store.Store,
	nats nats.Client,
	devClient deviceauth.Client,
	invClient inventory.Client,
) Indexer {
	return &indexer{
		store:     store,
		nats:      nats,
		devClient: devClient,
		invClient: invClient,
	}
}
