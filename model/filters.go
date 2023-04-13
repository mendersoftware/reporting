// Copyright 2023 Northern.tech AS
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
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

var validSelectors = []interface{}{
	"$eq",
	"$gt",
	"$gte",
	"$in",
	"$lt",
	"$lte",
	"$ne",
	"$nin",
	"$exists",
	"$regex",
}

const (
	SortOrderAsc  = "asc"
	SortOrderDesc = "desc"
)

var validSortOrders = []interface{}{SortOrderAsc, SortOrderDesc}

type SearchParams struct {
	Page                 int                   `json:"page"`
	PerPage              int                   `json:"per_page"`
	Filters              []FilterPredicate     `json:"filters"`
	GeoDistanceFilter    *GeoDistanceFilter    `json:"geo_distance_filter"`
	GeoBoundingBoxFilter *GeoBoundingBoxFilter `json:"geo_bounding_box_filter"`
	Sort                 []SortCriteria        `json:"sort"`
	Attributes           []SelectAttribute     `json:"attributes"`
	DeviceIDs            []string              `json:"device_ids"`
	Groups               []string              `json:"-"`
	TenantID             string                `json:"-"`
}

type FilterPredicate struct {
	Scope     string      `json:"scope" bson:"scope"`
	Attribute string      `json:"attribute" bson:"attribute"`
	Type      string      `json:"type" bson:"type"`
	Value     interface{} `json:"value" bson:"value"`
}

type GeoDistanceFilter struct {
	GeoDistance GeoDistance `json:"geo_distance" bson:"geo_distance"`
}

type GeoDistance struct {
	Distance string   `json:"distance" bson:"distance"`
	Location GeoPoint `json:"location" bson:"location"`
}

type GeoBoundingBoxFilter struct {
	GeoBoundingBox GeoBoundingBox `json:"geo_bounding_box" bson:"geo_bounding_box"`
}

type GeoBoundingBox struct {
	Location BoundingBox `json:"location" bson:"location"`
}

type BoundingBox struct {
	TopLeft     GeoPoint `json:"top_left" bson:"top_left"`
	BottomRight GeoPoint `json:"bottom_right" bson:"bottom_right"`
}

type GeoPoint struct {
	Lat float32 `json:"lat" bson:"lat"`
	Lon float32 `json:"lon" bson:"lon"`
}

type SortCriteria struct {
	Scope     string `json:"scope"`
	Attribute string `json:"attribute"`
	Order     string `json:"order"`
}

type SelectAttribute struct {
	Scope     string `json:"scope" bson:"scope"`
	Attribute string `json:"attribute" bson:"attribute"`
}

func (sp SearchParams) Validate() error {
	for _, f := range sp.Filters {
		err := f.Validate()
		if err != nil {
			return err
		}
	}

	for _, s := range sp.Sort {
		err := validation.ValidateStruct(&s,
			validation.Field(&s.Scope, validation.Required),
			validation.Field(&s.Attribute, validation.Required),
			validation.Field(&s.Order,
				validation.Required, validation.In(validSortOrders...),
			),
		)
		if err != nil {
			return err
		}
	}

	for _, s := range sp.Attributes {
		err := validation.ValidateStruct(&s,
			validation.Field(&s.Scope, validation.Required),
			validation.Field(&s.Attribute, validation.Required))
		if err != nil {
			return err
		}
	}
	return nil
}

func (f FilterPredicate) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Scope, validation.Required),
		validation.Field(&f.Attribute, validation.Required),
		validation.Field(&f.Type, validation.Required, validation.In(validSelectors...)),
		validation.Field(&f.Value, validation.NotNil))
}

// ValueType returns actual type info of the value:
// type, is_array, err
func (f FilterPredicate) ValueType() (Type, bool, error) {
	isArr := false
	typ := TypeStr

	switch f.Value.(type) {
	case bool:
		typ = TypeBool
	case float64:
		typ = TypeNum
	case string:
		break
	case []string:
		isArr = true
	case []interface{}:
		isArr = true
		ival := f.Value.([]interface{})
		switch ival[0].(type) {
		case bool:
			typ = TypeBool
		case float64:
			typ = TypeNum
		case string:
			break
		default:
			return 0, false, errors.New(
				fmt.Sprintf("unknown attribute value type: %v %T",
					ival[0], ival[0]),
			)
		}
	default:
		return 0, false, errors.New(
			fmt.Sprintf("unknown attribute value type: %v %T",
				f.Value, f.Value),
		)

	}

	return typ, isArr, nil
}
