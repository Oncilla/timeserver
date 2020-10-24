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

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

type claim struct {
	User string `json:"user"`
	Role Role   `json:"role"`
	jwt.StandardClaims
}

func (c claim) Valid() error {
	if err := c.StandardClaims.Valid(); err != nil {
		return err
	}
	vErr := new(jwt.ValidationError)
	if c.User == "" {
		vErr.Inner = fmt.Errorf("user is not set")
		vErr.Errors |= jwt.ValidationErrorMalformed
	}
	if c.Role == 0 {
		vErr.Inner = fmt.Errorf("invalid role")
		vErr.Errors |= jwt.ValidationErrorMalformed
	}
	if vErr.Errors != 0 {
		return vErr
	}
	return nil
}
