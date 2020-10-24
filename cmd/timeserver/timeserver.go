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

package main

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"

	"github.com/brpaz/echozap"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/oncilla/timeserver/pkg/api"
	"github.com/oncilla/timeserver/pkg/api/gen"
	"github.com/oncilla/timeserver/pkg/timeserver"
	"github.com/oncilla/timeserver/pkg/ui"
)

// CommandPather returns the path to a command.
type CommandPather interface {
	CommandPath() string
}

func main() {
	var flags struct {
		addr  string
		store string
	}
	cmd := &cobra.Command{
		Use:           "timeserver",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			s, err := api.NewKeyStore(flags.store)
			if err != nil {
				return err
			}
			var key [32]byte
			if _, err := rand.Read(key[:]); err != nil {
				return err
			}

			zapCfg := zap.NewProductionConfig()
			logger, err := zapCfg.Build()
			if err != nil {
				return err
			}
			swagger, err := gen.GetSwagger()
			if err != nil {
				return err
			}

			e := echo.New()
			e.Use(echozap.ZapLogger(logger))

			timeServer := &timeserver.Server{}
			e.GET("/", timeServer.Serve)

			e.GET("/spec", func(c echo.Context) error {
				return c.JSON(200, swagger)
			})
			e.GET("/ui", func(c echo.Context) error { return c.Redirect(301, "./ui/index.html") })
			e.GET("/ui/*", echo.WrapHandler(http.StripPrefix("/ui", http.FileServer(ui.Dir()))))

			apiServer := &api.Server{
				TimeZone: &timeServer.Zone,
				LogLevel: &zapCfg.Level,
				Logger:   logger,
				KeyStore: s,
				JWTKey:   key[:],
			}
			api.RegisterHandlers(e, apiServer)
			return e.Start("0.0.0.0:8081")
		},
	}
	cmd.AddCommand(
		newAPIKey(cmd),
		newCompletion(cmd),
		newVersion(cmd),
	)
	cmd.Flags().StringVar(&flags.addr, "addr", "localhost:8081", "Address to serve on")
	cmd.Flags().StringVar(&flags.store, "store", ".timeserver/store", "Path to the API key store")
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
