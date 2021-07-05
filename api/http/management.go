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
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/rest.utils"

	"github.com/mendersoftware/reporting/app/reporting"
	"github.com/mendersoftware/reporting/model"
)

const (
	hdrTotalCount = "X-Total-Count"
)

type ManagementController struct {
	reporting reporting.App
}

func NewManagementController(r reporting.App) *ManagementController {
	return &ManagementController{
		reporting: r,
	}
}

func (mc *ManagementController) Search(c *gin.Context) {
	params, err := parseSearchParams(c)

	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "malformed request body"),
		)
		return
	}

	ctx := c.Request.Context()

	id := identity.FromContext(ctx)
	if id == nil || id.Tenant == "" {
		rest.RenderError(c,
			http.StatusUnauthorized,
			errors.New("tenant claim not present in JWT"),
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

func parseSearchParams(c *gin.Context) (*model.SearchParams, error) {
	var searchParams model.SearchParams

	err := c.ShouldBindJSON(&searchParams)
	if err != nil {
		return nil, err
	}

	if searchParams.Page < 1 {
		searchParams.Page = 1
	}
	if searchParams.PerPage < 1 {
		searchParams.PerPage = 20
	}

	if err := searchParams.Validate(); err != nil {
		return nil, err
	}

	return &searchParams, nil
}

func pageLinkHdrs(c *gin.Context, page, perPage, total int) {
	url := &url.URL{
		Path:     c.Request.URL.Path,
		RawQuery: c.Request.URL.RawQuery,
		Fragment: c.Request.URL.Fragment,
	}

	query := url.Query()

	query.Set("page", "1")
	query.Set("per_page", fmt.Sprintf("%d", perPage))
	url.RawQuery = query.Encode()
	Link := fmt.Sprintf(`<%s>;rel="first"`, url.String())
	// Previous page
	if page > 1 {
		query.Set("page", fmt.Sprintf("%d", page-1))
		url.RawQuery = query.Encode()
		Link = fmt.Sprintf(`%s, <%s>;rel="previous"`, Link, url.String())
	}

	// Next page
	if total > (perPage*page - 1) {
		query.Set("page", fmt.Sprintf("%d", page+1))
		url.RawQuery = query.Encode()
		Link = fmt.Sprintf(`%s, <%s>;rel="next"`, Link, url.String())

	}
	c.Header("Link", Link)
}
