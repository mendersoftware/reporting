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

package http

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/mendersoftware/go-lib-micro/log"
)

// API URL used by the HTTP router
const (
	URIInternal   = "/api/internal/v1/reporting"
	URIManagement = "/api/management/v1/reporting"

	URILiveliness = "/health/alive"
)

// NewRouter returns the gin router
func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	router := gin.New()
	ctx := context.Background()
	l := log.FromContext(ctx)

	router.Use(routerLogger(l))
	router.Use(gin.Recovery())

	internal := NewInternalController()
	internalAPI := router.Group(URIInternal)
	internalAPI.GET(URILiveliness, internal.HealthAlive)

	return router
}
