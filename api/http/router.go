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

package http

import (
	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/go-lib-micro/accesslog"
	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/rbac"
	"github.com/mendersoftware/go-lib-micro/requestid"

	"github.com/mendersoftware/reporting/app/reporting"
)

// API URL used by the HTTP router
const (
	URIInternal   = "/api/internal/v1/reporting"
	URIManagement = "/api/management/v1/reporting"

	URIAlive                   = "/alive"
	URIHealth                  = "/health"
	URIInventoryAggregate      = "/devices/aggregate"
	URIInventoryAttrs          = "/devices/attributes"
	URIInventorySearch         = "/devices/search"
	URIInventorySearchAttrs    = "/devices/search/attributes"
	URIInventorySearchInternal = "/inventory/tenants/:tenant_id/search"
)

// NewRouter returns the gin router
func NewRouter(reporting reporting.App) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	router := gin.New()
	router.Use(accesslog.Middleware())
	router.Use(gin.Recovery())
	router.Use(requestid.Middleware())

	internal := NewInternalController(reporting)
	internalAPI := router.Group(URIInternal)
	internalAPI.GET(URIAlive, internal.Alive)
	internalAPI.GET(URIHealth, internal.Health)
	internalAPI.POST(URIInventorySearchInternal, internal.Search)

	mgmt := NewManagementController(reporting)
	mgmtAPI := router.Group(URIManagement)
	mgmtAPI.Use(identity.Middleware())
	mgmtAPI.Use(rbac.Middleware())
	mgmtAPI.POST(URIInventoryAggregate, mgmt.Aggregate)
	mgmtAPI.GET(URIInventoryAttrs, mgmt.Attrs)
	mgmtAPI.POST(URIInventorySearch, mgmt.Search)
	mgmtAPI.GET(URIInventorySearchAttrs, mgmt.SearchAttrs)

	return router
}
