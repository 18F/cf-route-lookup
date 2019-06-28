package main

import (
	"flag"
	"log"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
)

var CMD = "lookup-route"

type BasicPlugin struct{}

// possibleDomains returns all domain levels, down to the second-level domain (SLD), in order.
func getPossibleDomains(hostname string) []string {
	parts := strings.Split(hostname, ".")
	numCombinations := len(parts) - 1
	possibleDomains := make([]string, numCombinations)
	for i := 0; i < numCombinations; i++ {
		possibleDomains[i] = strings.Join(parts[i:], ".")
	}
	return possibleDomains
}

func (c *BasicPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	log.SetFlags(0)

	if len(args) == 1 && args[0] == "CLI-MESSAGE-UNINSTALL" {
		log.Println("Lack of features? Please contribute: https://github.com/18F/cf-route-lookup/")
		return
	}

	if len(args) >= 1 && args[0] != CMD {
		log.Println("Type --help...")
		return
	}

	flags := flag.NewFlagSet(CMD, flag.ContinueOnError)
	target := flags.Bool("t", false, "Target the org / space containing this route")
	_ = flags.Parse(args[1:])

	if len(flags.Args()) != 1 {
		log.Fatal("Please specify the domain to look up.")
	}

	hostname := flags.Args()[0]

	appsExt, err := getApps(cliConnection, hostname)
	if err != nil {
		log.Fatal("Error retrieving apps: ", err)
	}

	if len(appsExt) == 0 {
		log.Println("Not bound to any applications.")
		return
	}
	log.Println("Bound to:")

	var firstApp App
	found := false

	sort.Slice(appsExt, func(i, j int) bool {
		return appsExt[i].route.route.SpaceGUID < appsExt[j].route.route.SpaceGUID
	})

	targets := NewTargets()
	spaceGuid := ""
	for _, appExt := range appsExt {
		if appExt.app != (App{}) {
			if !found {
				firstApp = appExt.app
				found = true
			}
		}

		if spaceGuid != appExt.route.route.SpaceGUID {
			spaceGuid = appExt.route.route.SpaceGUID
			org, space, err := targets.GetTargetBySpaceGuid(cliConnection, spaceGuid)
			if err != nil {
				log.Println("Cannot find the space with GUID=" + spaceGuid)
				return
			}
			log.Printf("\n> cf target -o %v -s %v\n", org.Entity.Name, space.Entity.Name)
		}

		if appExt.app == (App{}) {
			log.Printf("  > # unbounded route: %v (%v)\n", appExt.route.hostname, appExt.route.route.GUID)
		} else {
			log.Printf("  > cf app %v  # route: %v (%v)\n", appExt.app.Entity.Name, appExt.route.hostname, appExt.route.route.GUID)
		}
	}

	if *target {
		firstAppOrg, firstAppSpace, err := targets.GetTargetBySpaceGuid(cliConnection, firstApp.Entity.SpaceGUID)
		if err != nil {
			log.Fatal("Error retrieving target: ", err)
		}
		log.Printf("\nChanging target to: %v ...", firstAppOrg.Entity.Name+" / "+firstAppSpace.Entity.Name)
		_, err = cliConnection.CliCommand("target", "-o", firstAppOrg.Entity.Name, "-s", firstAppSpace.Entity.Name)
		if err != nil {
			log.Fatal("Error targeting app: ", err)
		}
	}

}

func (c *BasicPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "route-lookup",
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
				Name:     CMD,
				HelpText: "Look up the mapping of a provided route." +
					"\n   Wildcard support: '*' - any number of any characters with the exception of dots",
				UsageDetails: plugin.Usage{
					Usage: "\ncf " + CMD + " [-t] some.example.com" +
						"\ncf " + CMD + " [-t] *.example.com" +
						"\ncf " + CMD + " [-t] *s*me*.example.com",
					Options: map[string]string{
						"t": "Target the first org / space matching the route",
					},
				},
			},
		},
	}
}

func main() {
	plugin.Start(new(BasicPlugin))
}
