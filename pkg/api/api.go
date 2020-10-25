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
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/oncilla/timeserver/pkg/api/gen"
	"github.com/oncilla/timeserver/pkg/timeserver"
)

const (
	apiKeyInfo = "api-key-info"
	apiToken   = "api-token"
)

// Server implements the configuration API.
type Server struct {
	LogLevel *zap.AtomicLevel
	TimeZone *timeserver.TimeZone
	KeyStore *KeyStore
	Logger   *zap.Logger
	JWTKey   []byte
}

// RegisterHandlers registers all the API handlers.
func RegisterHandlers(e *echo.Echo, s *Server) {
	w := gen.ServerInterfaceWrapper{
		Handler: s,
	}

	e.GET("/api/token", w.GetToken, s.apiKeyAuth(ConfigReader, ConfigWriter, Admin))

	configReaderGroup := e.Group("/api", s.withPrivilege(ConfigReader, ConfigWriter, Admin))
	configReaderGroup.GET("/log/level", s.GetLogLevel)
	configReaderGroup.GET("/timezone", s.GetTimeZone)

	configWriterGroup := e.Group("/api", s.withPrivilege(ConfigWriter, Admin))
	configWriterGroup.PUT("/log/level", s.SetLogLevel)
	configWriterGroup.PUT("/timezone", s.SetTimeZone)

	adminGroup := e.Group("/api/admin", s.withPrivilege(Admin))
	adminGroup.GET("/keys", w.AdminGetKeys)
	adminGroup.PUT("/keys", w.AdminSetKey)
	adminGroup.DELETE("/keys/:apiKeyID", w.AdminDeleteKey)
	adminGroup.GET("/keys/:apiKeyID", w.AdminGetKey)
}

// GetLogLevel wraps the zap log level handler
func (s *Server) GetLogLevel(ctx echo.Context) error {
	s.LogLevel.ServeHTTP(ctx.Response(), ctx.Request())
	return nil
}

// SetLogLevel wraps the zap log level handler
func (s *Server) SetLogLevel(ctx echo.Context) error {
	s.LogLevel.ServeHTTP(ctx.Response(), ctx.Request())
	return nil
}

// GetTimeZone serves the currently configured time zone.
func (s *Server) GetTimeZone(ctx echo.Context) error {
	return ctx.JSON(200, gen.TimeZone{Timezone: s.TimeZone.Get().String()})
}

// SetTimeZone sets the time zone.
func (s *Server) SetTimeZone(ctx echo.Context) error {
	var tz gen.TimeZone
	if err := ctx.Bind(&tz); err != nil {
		return err
	}
	if tz.Timezone == "" {
		return echo.ErrBadRequest
	}
	loc, err := time.LoadLocation(tz.Timezone)
	if err != nil {
		return err
	}
	s.TimeZone.Set(loc)
	return s.GetTimeZone(ctx)
}

// GetToken issues a token.
func (s *Server) GetToken(ctx echo.Context, params gen.GetTokenParams) error {
	info, ok := ctx.Get(apiKeyInfo).(Info)
	if !ok {
		return echo.ErrBadRequest
	}
	role := info.Role
	if params.Role != nil {
		var requested Role
		if err := requested.UnmarshalText([]byte(*params.Role)); err != nil {
			return err
		}
		if role < requested {
			return echo.ErrUnauthorized
		}
		role = requested
	}
	token := jwt.New(jwt.SigningMethodHS256)
	c := claim{
		User: info.User,
		Role: role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(6 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token.Claims = c
	t, err := token.SignedString(s.JWTKey)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, &gen.Token{
		ExpiresAt: int(c.ExpiresAt),
		Role:      c.Role.String(),
		Token:     t,
	})
}

// AdminGetKeys serves a list of a API key infos.
func (s *Server) AdminGetKeys(ctx echo.Context) error {
	infos, err := s.KeyStore.All()
	if err != nil {
		return err
	}
	rep := make(gen.APIKeys, 0, len(infos))
	for _, info := range infos {
		rep = append(rep, gen.APIKey{
			Id:   info.ID,
			Role: info.Role.String(),
			User: info.User,
		})
	}
	return ctx.JSON(200, rep)
}

// AdminSetKey sets a API key.
func (s *Server) AdminSetKey(ctx echo.Context) error {
	var key gen.APIKey
	if err := ctx.Bind(&key); err != nil {
		return err
	}
	var role Role
	if err := role.UnmarshalText([]byte(key.Role)); err != nil {
		return echo.ErrBadRequest
	}
	err := s.KeyStore.Add(key.Id, Info{
		ID:   key.Id,
		User: key.User,
		Role: role,
	})
	if err != nil {
		return err
	}
	return ctx.JSON(200, gen.Success{
		Message: fmt.Sprintf("created API key: user=%q role=%v", key.User, role),
	})
}

// AdminDeleteKey deletes an API key.
func (s *Server) AdminDeleteKey(ctx echo.Context, apiKeyID string) error {
	deleted, err := s.KeyStore.Delete(apiKeyID)
	if err != nil {
		s.Logger.Error(err.Error())
		return err
	}
	if !deleted {
		return echo.ErrNotFound
	}
	return ctx.JSON(200, gen.Success{
		Message: "API key deleted",
	})
}

// AdminGetKey serves a single API key info.
func (s *Server) AdminGetKey(ctx echo.Context, apiKeyID string) error {
	info, err := s.KeyStore.Get(apiKeyID)
	if err != nil {
		return err
	}
	if info == nil {
		return ctx.JSON(404, gen.Error{
			Error: "API key not found",
		})
	}
	return ctx.JSON(200, gen.APIKey{
		Id:   apiKeyID,
		Role: info.Role.String(),
		User: info.User,
	})
}

func (s *Server) withPrivilege(roles ...Role) echo.MiddlewareFunc {
	apiAuth := s.apiKeyAuth(roles...)

	jwtConfig := middleware.JWTConfig{
		Claims:     &claim{},
		SigningKey: s.JWTKey,
		ContextKey: apiToken,
	}
	jwtAuth := middleware.JWTWithConfig(jwtConfig)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if len(c.Request().Header.Get("X-API-KEY")) != 0 {
				return apiAuth(next)(c)
			}
			return jwtAuth(func(c echo.Context) error {
				user := c.Get(apiToken).(*jwt.Token)
				claims := user.Claims.(*claim)
				for _, role := range roles {
					if role == claims.Role {
						c.Set(apiKeyInfo, Info{
							User: claims.User,
							Role: claims.Role,
						})
						return next(c)
					}
				}
				return echo.ErrUnauthorized
			})(c)
		}
	}
}

func (s *Server) apiKeyAuth(roles ...Role) echo.MiddlewareFunc {
	apiKeyConfig := middleware.KeyAuthConfig{
		KeyLookup: "header:X-API-KEY",
		Validator: func(apiKey string, ctx echo.Context) (bool, error) {
			if len(roles) == 0 {
				return true, nil
			}
			info, err := s.KeyStore.Get(apiKey)
			if info == nil || err != nil {
				return false, err
			}
			for _, role := range roles {
				if role == info.Role {
					ctx.Set(apiKeyInfo, *info)
					return true, nil
				}
			}
			return false, nil
		},
	}
	return middleware.KeyAuthWithConfig(apiKeyConfig)
}
