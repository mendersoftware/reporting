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
	"github.com/pkg/errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/rest.utils"

	"github.com/mendersoftware/reporting/app/reporting"
)

// InternalController contains internal end-points
type InternalController struct {
	reporting reporting.App
}

// NewInternalController returns a new InternalController
func NewInternalController(r reporting.App) *InternalController {
	return &InternalController{
		reporting: r,
	}
}

// Alive responds to GET /health/alive
func (h InternalController) Alive(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

func (mc *InternalController) Search(c *gin.Context) {
	tid := c.Param("tenant_id")

	ctx := c.Request.Context()

	ctx = identity.WithContext(ctx, &identity.Identity{Tenant: tid})

	params, err := parseSearchParams(c)

	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "malformed request body"),
		)
		return
	}

	res, total, err := mc.reporting.InventorySearchDevices(ctx, params)
	if err != nil {
		rest.RenderError(c,
			http.StatusInternalServerError,
			err,
		)
		return
	}

	pageLinkHdrs(c, params.Page, params.PerPage, total)

	c.Header(hdrTotalCount, strconv.Itoa(total))
	c.JSON(http.StatusOK, res)
}

func (ic *InternalController) Reindex(c *gin.Context) {
	tid := c.Param("tenant_id")
	did := c.Param("device_id")

	service := c.Query("service")

	ctx := c.Request.Context()
	ctx = identity.WithContext(ctx, &identity.Identity{Tenant: tid})

	err := ic.reporting.Reindex(ctx, tid, did, service)

	switch err {
	case nil:
		c.Status(http.StatusAccepted)
	case reporting.ErrUnknownService:
		if err != nil {
			rest.RenderError(c,
				http.StatusBadRequest,
				err,
			)
			return
		}
	default:
		rest.RenderError(c,
			http.StatusInternalServerError,
			err,
		)
		return
	}
}
