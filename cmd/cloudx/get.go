// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cloudx

import (
	"github.com/spf13/cobra"

	"github.com/ory/x/cmdx"

	"github.com/ory/cli/cmd/cloudx/client"
	"github.com/ory/cli/cmd/cloudx/identity"
	"github.com/ory/cli/cmd/cloudx/oauth2"
	"github.com/ory/cli/cmd/cloudx/project"
	"github.com/ory/cli/cmd/cloudx/workspace"
)

func NewGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a resource",
	}

	cmd.AddCommand(
		project.NewGetProjectCmd(),
		project.NewGetKratosConfigCmd(),
		project.NewGetKetoConfigCmd(),
		project.NewGetOAuth2ConfigCmd(),
		workspace.NewGetCmd(),
		identity.NewGetIdentityCmd(),
		oauth2.NewGetOAuth2Client(),
		oauth2.NewGetJWK(),
	)

	client.RegisterConfigFlag(cmd.PersistentFlags())
	client.RegisterYesFlag(cmd.PersistentFlags())
	cmdx.RegisterNoiseFlags(cmd.PersistentFlags())

	return cmd
}
