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

// common enum for some type introspections we'll need
type Type int

const (
	TypeAny Type = iota
	TypeStr
	TypeNum
	TypeBool
)

// scope prefixes
const (
	scopeInventory = "inventory"
	scopeIdentity  = "identity"
	scopeSystem    = "system"
	scopeTags      = "tags"
	scopeMonitor   = "monitor"
)

// type enum/suffixes
const (
	typeStr  = "str"
	typeNum  = "num"
	typeBool = "bool"
)

var (
	attrSuffixes = map[Type]string{
		TypeStr:  typeStr,
		TypeNum:  typeNum,
		TypeBool: typeBool,
	}
)

// toAttr composes the flat-style attribute name based on
// scope, name, and type
func ToAttr(scope, name string, typ Type) string {
	return scope + "_" + Dedot(name) + "_" + attrSuffixes[typ]
}
