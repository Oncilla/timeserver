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
	"fmt"

	"github.com/oncilla/boa/pkg/boa"
	"github.com/spf13/cobra"

	"github.com/oncilla/timeserver/pkg/api"
)

func newAPIKey(pather CommandPather) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "api-key",
		Short: "Manage API keys",
	}
	cmd.AddCommand(newAPIKeyAdd(boa.Pather(pather.CommandPath() + " api-key")))
	return cmd
}

func newAPIKeyAdd(pather CommandPather) *cobra.Command {
	var flags struct {
		store string
		role  string
		user  string
	}
	var cmd = &cobra.Command{
		Use:     "add <API-key>",
		Short:   "Add the API key to the store",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf("  %[1]s add <API-key>", pather.CommandPath()),
		RunE: func(cmd *cobra.Command, args []string) error {
			var role api.Role
			if err := role.UnmarshalText([]byte(flags.role)); err != nil {
				return fmt.Errorf("parsing role: %w", err)
			}
			if flags.user == "" {
				return fmt.Errorf("user must be set")
			}
			cmd.SilenceUsage = true

			s, err := api.NewKeyStore(flags.store)
			if err != nil {
				return fmt.Errorf("opening store: %w", err)
			}

			info := api.Info{
				ID:   args[0],
				User: flags.user,
				Role: role,
			}
			if err := s.Add(args[0], info); err != nil {
				return err
			}

			fmt.Printf("Added API key: user=%q role=%s\n", info.User, info.Role)
			return nil
		},
	}
	cmd.Flags().StringVar(&flags.store, "store", ".timeserver/store", "Path to the API key store")
	cmd.Flags().StringVar(&flags.role, "role", "config:reader", "Role associated with the API key")
	cmd.Flags().StringVar(&flags.user, "user", "", "User associated with the API key")
	cmd.MarkFlagRequired("user")
	return cmd
}
