package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/plugin"
)

// BasicPlugin is the struct implementing the interface defined by the core CLI. It can
// be found at  "code.cloudfoundry.org/cli/plugin/plugin.go"
type BasicPlugin struct{}

type RoutesResponse struct {
	NextUrl   string       `json:"next_url"`
	Resources []ccv2.Route `json:"resources"`
}

func apiCall(cliConnection plugin.CliConnection, path string) (body string, err error) {
	// based on https://github.com/krujos/cfcurl/blob/320854091a119f220102ba356e507c361562b221/cfcurl.go
	bodyLines, err := cliConnection.CliCommandWithoutTerminalOutput("curl", path)
	if err != nil {
		return
	}
	body = strings.Join(bodyLines, "\n")
	return
}

func getRoutes(cliConnection plugin.CliConnection) (routes []ccv2.Route, err error) {
	// based on https://github.com/ECSTeam/buildpack-usage/blob/e2f7845f96c021fa7f59d750adfa2f02809e2839/command/buildpack_usage_cmd.go#L161-L167

	routes = make([]ccv2.Route, 0)
	url := "/v2/routes?results-per-page=100"

	// paginate
	for url != "" {
		var body string
		body, err = apiCall(cliConnection, url)
		if err != nil {
			return
		}

		var data RoutesResponse
		err = json.Unmarshal([]byte(body), &data)
		if err != nil {
			return
		}

		routes = append(routes, data.Resources...)
		url = data.NextUrl
	}

	return
}

// Run must be implemented by any plugin because it is part of the
// plugin interface defined by the core CLI.
//
// Run(....) is the entry point when the core CLI is invoking a command defined
// by a plugin. The first parameter, plugin.CliConnection, is a struct that can
// be used to invoke cli commands. The second paramter, args, is a slice of
// strings. args[0] will be the name of the command, and will be followed by
// any additional arguments a cli user typed in.
//
// Any error handling should be handled with the plugin itself (this means printing
// user facing errors). The CLI will exit 0 if the plugin exits 0 and will exit
// 1 should the plugin exits nonzero.
func (c *BasicPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	// Ensure that we called the command basic-plugin-command
	if args[0] == "basic-plugin-command" {
		fmt.Println("Running the basic-plugin-command")

		// TODO check for argument length

		routes, err := getRoutes(cliConnection)
		if err != nil {
			log.Fatal("Error retrieving the routes.")
		}
		fmt.Println(len(routes), "routes found.")

		subdomain := strings.Split(args[1], ".")[0]
		matches := make([]ccv2.Route, 0, len(routes))
		for _, route := range routes {
			// TODO handle private domains, which may not have a Host
			if route.Host == subdomain {
				fmt.Println("Subdomain match!", subdomain)
				matches = append(matches, route)
			}
		}
		if len(matches) == 0 {
			fmt.Println("Domain not found.")
		}
	}
}

// GetMetadata must be implemented as part of the plugin interface
// defined by the core CLI.
//
// GetMetadata() returns a PluginMetadata struct. The first field, Name,
// determines the name of the plugin which should generally be without spaces.
// If there are spaces in the name a user will need to properly quote the name
// during uninstall otherwise the name will be treated as seperate arguments.
// The second value is a slice of Command structs. Our slice only contains one
// Command Struct, but could contain any number of them. The first field Name
// defines the command `cf basic-plugin-command` once installed into the CLI. The
// second field, HelpText, is used by the core CLI to display help information
// to the user in the core commands `cf help`, `cf`, or `cf -h`.
func (c *BasicPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "MyBasicPlugin",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 0,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 7,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "basic-plugin-command",
				HelpText: "Basic plugin command's help text",

				// UsageDetails is optional
				// It is used to show help of usage of each command
				UsageDetails: plugin.Usage{
					Usage: "basic-plugin-command\n   cf basic-plugin-command",
				},
			},
		},
	}
}

// Unlike most Go programs, the `Main()` function will not be used to run all of the
// commands provided in your plugin. Main will be used to initialize the plugin
// process, as well as any dependencies you might require for your
// plugin.
func main() {
	// Any initialization for your plugin can be handled here
	//
	// Note: to run the plugin.Start method, we pass in a pointer to the struct
	// implementing the interface defined at "code.cloudfoundry.org/cli/plugin/plugin.go"
	//
	// Note: The plugin's main() method is invoked at install time to collect
	// metadata. The plugin will exit 0 and the Run([]string) method will not be
	// invoked.
	plugin.Start(new(BasicPlugin))
	// Plugin code should be written in the Run([]string) method,
	// ensuring the plugin environment is bootstrapped.
}
