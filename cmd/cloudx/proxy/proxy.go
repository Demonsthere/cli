// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"fmt"
	"net/url"

	"github.com/ory/cli/cmd/cloudx/client"
	"github.com/ory/x/cmdx"

	"github.com/spf13/cobra"
)

func NewProxyCommand() *cobra.Command {
	conf := config{
		pathPrefix: "/.ory",
	}

	proxyCmd := &cobra.Command{
		Use:   "proxy <application-url> [<publish-url>]",
		Short: "Run your app and Ory on the same domain using a reverse proxy",
		Args:  cobra.RangeArgs(1, 2),
		Example: `{{.CommandPath}} http://localhost:3000 --dev
{{.CommandPath}} http://localhost:3000 https://app.example.com \
	--allowed-cors-origins https://www.example.org \
	--allowed-cors-origins https://api.example.org \
	--allowed-cors-origins https://www.another-app.com
`,
		Long: `The Ory Proxy allows your application and Ory to run on the same domain by acting as a reverse proxy. It forwards all traffic to your application, ensuring that features like cookies and CORS function correctly during local development.

The first argument, ` + "`application-url`" + `, points to the location of your application. The Ory Proxy will pass all traffic through to this URL.

Example usage:

		$ {{.CommandPath}} --project <project-id-or-slug> https://www.example.org
		$ ORY_PROJECT=<project-id-or-slug> {{.CommandPath}} proxy http://localhost:3000

### Connecting to Ory

Before using the Ory Proxy, you need to have an Ory Network project. You can create a new project with the following command:

		$ {{.Root.Name}} create project --name "Command Line Project"

Once your project is ready, pass the project’s slug to the proxy command:

		$ {{.CommandPath}} --project <project-id-or-slug> ...

### Local development

For local development, use the ` + "`--dev`" + ` flag to apply a relaxed security setting:

		$ {{.CommandPath}} --dev --project <project-id-or-slug> http://localhost:3000

The first argument, ` + "`application-url`" + `, points to your application's location. If running both the proxy and your app on the same host, this could be ` + "`localhost`" + `. All traffic sent to the Ory Proxy will be forwarded to this URL.

The second argument, ` + "`publish-url`" + `, is optional and only necessary for production scenarios. It specifies the public URL of your application (e.g., ` + "`https://www.example.org`" + `). If ` + "`publish-url`" + ` is not set, it defaults to the host and port the proxy listens on.

**Important**: The Ory Proxy is intended for development use only and should not be used in production environments.

### Connecting in automated environments

To connect the Ory Tunnel in automated environments, create a Project API Key for your project and set it as an environment variable:

		$ %[2]s=<project-api-key> {{.CommandPath}} tunnel ...

This will prevent the browser window from opening.

### Running behind a gateway (development only)

If you are using the Ory Proxy behind a gateway during development, you must set the ` + "`publish-url`" + ` argument:

		$ {{.CommandPath}} --project <project-id-or-slug> \
		  http://localhost:3000 \
		  https://gateway.local:5000

Note: You cannot set a path in the ` + "`publish-url`" + `.

### Ports

By default, the proxy listens on port 4000. To change this, use the ` + "`--port`" + ` flag:

		$ {{.CommandPath}} --port 8080 --project <project-id-or-slug> http://localhost:3000

### Multiple domains

If the proxy runs on a subdomain and you want Ory’s cookies (e.g., session cookies) to be accessible across all your domains, use the ` + "`--cookie-domain`" + ` flag to customize the cookie domain. Additionally, allow your subdomains in the CORS headers:

		$ {{.CommandPath}} --project <project-id-or-slug> \
		  --cookie-domain gateway.local \
		  --allowed-cors-origins https://www.gateway.local \
		  --allowed-cors-origins https://api.gateway.local \
		  http://127.0.0.1:3000 \
		  https://ory.gateway.local

### Redirects

By default, all redirects will point to ` + "`publish-url`" + `. You can customize this behavior using the ` + "`--default-redirect-url`" + ` flag:

		$ {{.CommandPath}} --project <project-id-or-slug> \
		  --default-redirect-url /welcome \
		  http://127.0.0.1:3000 \
		  https://ory.example.org

This ensures that all redirects (e.g., after login) go to ` + "`/welcome`" + ` instead of ` + "`/`" + `, unless you’ve specified custom redirects in your Ory configuration or via the flow’s ` + "`?return_to=`" + ` query parameter.

### JSON Web Token

When a request is not authenticated, the HTTP ` + "`Authorization`" + ` header will be empty:

		GET / HTTP/1.1
		Host: localhost:3000

If the request is authenticated, a JSON Web Token (JWT) containing the Ory session will be sent in the HTTP ` + "`Authorization`" + ` header:

		GET / HTTP/1.1
		Host: localhost:3000
		Authorization: Bearer the-json-web-token

The JWT claims contain:
- The ` + "`sub`" + ` field, which is set to the Ory Identity ID.
- The ` + "`session`" + ` field, which contains the full Ory Session.

The JWT is signed using the ES256 algorithm. You can fetch the public key by querying the ` + "`/ory/jwks.json`" + ` endpoint, for example:

http://127.0.0.1:4000/.ory/jwks.json

An example JWT payload:

		{
		  "id": "821f5a53-a0b3-41fa-9c62-764560fa4406",
		  "active": true,
		  "expires_at": "2021-02-25T09:25:37.929792Z",
		  "authenticated_at": "2021-02-24T09:25:37.931774Z",
		  "issued_at": "2021-02-24T09:25:37.929813Z",
		  "identity": {
			"id": "18aafd3e-b00c-4b19-81c8-351e38705126",
			"schema_id": "default",
			"schema_url": "https://example.projects.oryapis.com/api/kratos/public/schemas/default",
			"traits": {
			  "email": "foo@bar"
			  // ... other identity traits
			}
		  }
		}
`,

		RunE: func(cmd *cobra.Command, args []string) error {
			conf.upstream = args[0]

			selfURLString := fmt.Sprintf("http://localhost:%d", conf.port)
			if len(args) == 2 {
				selfURLString = args[1]
			}

			var err error
			conf.publicURL, err = url.ParseRequestURI(selfURLString)
			if err != nil {
				return err
			}

			if conf.defaultRedirectTo.String() == "" {
				conf.defaultRedirectTo.URL = *conf.publicURL
			}

			h, err := client.NewCobraCommandHelper(cmd)
			if err != nil {
				return err
			}

			return runReverseProxy(cmd.Context(), h, cmd.ErrOrStderr(), &conf, "proxy")
		},
	}

	flags := proxyCmd.Flags()
	registerConfigFlags(&conf, flags)
	registerProxyConfigFlags(&conf, flags)

	client.RegisterConfigFlag(flags)
	client.RegisterProjectFlag(flags)
	client.RegisterWorkspaceFlag(flags)
	client.RegisterYesFlag(flags)
	cmdx.RegisterNoiseFlags(flags)

	proxyCmd.Root().Name()
	return proxyCmd
}
