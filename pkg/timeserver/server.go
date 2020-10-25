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

package timeserver

import (
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// Server is a simple time server where the time zone can be dynamically set.
type Server struct {
	Zone TimeZone
}

// Serve serves the time as JSON.
func (s *Server) Serve(c echo.Context) error {
	return c.JSON(200, time.Now().In(s.Zone.Get()))
}

// TimeZone is a serializing wrapper for a time zone.
type TimeZone struct {
	mu  sync.Mutex
	loc *time.Location
}

// Set sets the time zone.
func (z *TimeZone) Set(loc *time.Location) {
	z.mu.Lock()
	defer z.mu.Unlock()
	z.loc = loc
}

// Get returns the time zone. If it is nil, time.UTC is returned.
func (z *TimeZone) Get() *time.Location {
	z.mu.Lock()
	defer z.mu.Unlock()
	if z.loc == nil {
		return time.UTC
	}
	return z.loc
}
