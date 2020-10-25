// Copyright 2020 oncilla
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import "fmt"

// Role is the role of a user.
type Role int

// All roles.
const (
	UnknownRole Role = iota
	ConfigReader
	ConfigWriter
	Admin
)

func (r *Role) UnmarshalText(b []byte) error {
	switch string(b) {
	case ConfigReader.String():
		*r = ConfigReader
	case ConfigWriter.String():
		*r = ConfigWriter
	case Admin.String():
		*r = Admin
	default:
		return fmt.Errorf("unknown role: %s", string(b))
	}
	return nil
}

func (r Role) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r Role) String() string {
	switch r {
	case ConfigReader:
		return "config:reader"
	case ConfigWriter:
		return "config:writer"
	case Admin:
		return "admin"
	default:
		return "unknown"
	}
}
