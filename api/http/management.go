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
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/rbac"
	"github.com/mendersoftware/go-lib-micro/rest.utils"

	"github.com/mendersoftware/reporting/app/reporting"
	"github.com/mendersoftware/reporting/model"
)

const (
	ParamPageDefault    = 1
	ParamPerPageDefault = 20

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

type attributes struct {
	Limit      int         `json:"limit"`
	Count      int         `json:"count"`
	Attributes []attribute `json:"attributes"`
}

type attribute struct {
	Name  string `json:"name"`
	Scope string `json:"scope"`
}

func (mc *ManagementController) Attrs(c *gin.Context) {
	ctx := c.Request.Context()

	var tenantID string
	if id := identity.FromContext(ctx); id != nil && id.Tenant != "" {
		tenantID = id.Tenant
	} else {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.New("missing tenant ID from the context"),
		)
		return
	}

	mapping, err := mc.reporting.GetMapping(ctx, tenantID)
	if err != nil {
		rest.RenderError(c,
			http.StatusInternalServerError,
			errors.Wrap(err, "failed to retrieve the mapping"),
		)
		return
	} else if mapping == nil {
		mapping = &model.Mapping{}
	}

	attributesList := make([]attribute, 0, len(mapping.Inventory))
	for _, attr := range mapping.Inventory {
		parts := strings.SplitN(attr, string(os.PathSeparator), 2)
		attributesList = append(attributesList, attribute{
			Name:  parts[1],
			Scope: parts[0],
		})
	}
	res := &attributes{
		Limit:      model.MaxMappingInventoryAttributes,
		Count:      len(mapping.Inventory),
		Attributes: attributesList,
	}

	c.JSON(http.StatusOK, res)
}

func (mc *ManagementController) Search(c *gin.Context) {
	ctx := c.Request.Context()
	params, err := parseSearchParams(ctx, c)
	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "malformed request body"),
		)
		return
	}

	if scope := rbac.ExtractScopeFromHeader(c.Request); scope != nil {
		params.Groups = scope.DeviceGroups
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

func parseSearchParams(ctx context.Context, c *gin.Context) (*model.SearchParams, error) {
	var searchParams model.SearchParams

	err := c.ShouldBindJSON(&searchParams)
	if err != nil {
		return nil, err
	}

	if id := identity.FromContext(ctx); id != nil {
		searchParams.TenantID = id.Tenant
	} else {
		return nil, errors.New("missing tenant ID from the context")
	}

	if searchParams.PerPage <= 0 {
		searchParams.PerPage = ParamPerPageDefault
	}
	if searchParams.Page <= 0 {
		searchParams.Page = ParamPageDefault
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

func (mc *ManagementController) SearchAttrs(c *gin.Context) {
	ctx := c.Request.Context()

	id := identity.FromContext(ctx)
	res, err := mc.reporting.GetSearchableInvAttrs(ctx, id.Tenant)
	if err != nil {
		rest.RenderError(c,
			http.StatusInternalServerError,
			err,
		)
		return
	}

	c.JSON(http.StatusOK, res)
}
