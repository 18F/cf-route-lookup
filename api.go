package main

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/plugin"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strings"
)


func apiCall(cliConnection plugin.CliConnection, path string) (body string, err error) {
	/*
	   [{    "description": "Unknown request",    "error_code": "CF-NotFound",    "code": 10000 }]
	*/
	type ErrorResponse struct {
		Description string `json:"description"`
		ErrorCode string `json:"error_code"`
		Code int `json:"code"`
	}
	// based on https://github.com/krujos/cfcurl/blob/320854091a119f220102ba356e507c361562b221/cfcurl.go
	bodyLines, err := cliConnection.CliCommandWithoutTerminalOutput("curl", path)
	if err != nil {
		return "", err
	}
	if nil == bodyLines || 0 == len(bodyLines) {
		return "", errors.New("CF API returned no output")
	}
	body = strings.Join(bodyLines, "\n")
	var erResp ErrorResponse
	err = json.Unmarshal([]byte(body), &erResp)
	if err == nil && erResp.ErrorCode!=""{
		return "", errors.New("CF API ("+path+") returned error: ["+erResp.ErrorCode+"] "+erResp.Description)
	}
	return
}

func inQuery(filter string, values []string) string {
	return filter + " IN " + strings.Join(values, ",")
}

func getDomains(cliConnection plugin.CliConnection, names []string) (domains []ccv2.Domain, err error) {
	// based on https://github.com/ECSTeam/buildpack-usage/blob/e2f7845f96c021fa7f59d750adfa2f02809e2839/command/buildpack_usage_cmd.go#L161-L167

	domains = make([]ccv2.Domain, 0)

	endpoints := [...]string{"/v2/private_domains", "/v2/shared_domains"}

	params := url.Values{}
	params.Set("q", inQuery("name", names))
	params.Set("results-per-page", "100")
	queryString := params.Encode()

	for _, endpoint := range endpoints {
		uri := endpoint + "?" + queryString

		// paginate
		for uri != "" {
			var body string
			body, err = apiCall(cliConnection, uri)
			if err != nil {
				return
			}

			var data DomainsResponse
			err = json.Unmarshal([]byte(body), &data)
			if err != nil {
				return
			}

			domains = append(domains, data.Resources...)
			uri = data.NextUrl
		}
	}

	return
}

func getDomain(cliConnection plugin.CliConnection, hostname string) (matchingDomain ccv2.Domain, found bool, err error) {
	possibleDomains := getPossibleDomains(hostname)
	domains, err := getDomains(cliConnection, possibleDomains)
	if err != nil {
		return
	}

	for _, possibleDomain := range possibleDomains {
		for _, domain := range domains {
			if domain.Name == possibleDomain {
				found = true
				matchingDomain = domain
				return
			}
		}
	}
	return
}

func getDomainRoutes(cliConnection plugin.CliConnection, domain ccv2.Domain) (routes []ccv2.Route, err error) {
	// based on https://github.com/ECSTeam/buildpack-usage/blob/e2f7845f96c021fa7f59d750adfa2f02809e2839/command/buildpack_usage_cmd.go#L161-L167
	routes = make([]ccv2.Route, 0)
	params := url.Values{}
	// TODO also filter by host
	params.Set("q", "domain_guid:"+domain.GUID)
	params.Set("results-per-page", "100")
	uri := "/v2/routes?" + params.Encode()

	// paginate
	for uri != "" {
		var body string
		body, err = apiCall(cliConnection, uri)
		if err != nil {
			return
		}
		var data RoutesResponse
		err = json.Unmarshal([]byte(body), &data)
		if err != nil {
			return
		}
		routes = append(routes, data.Resources...)
		uri = data.NextUrl
	}

	return
}

func matchWildcard(wildcard string, route string) (bool, error) {
	wdRegex := strings.Replace(wildcard, ".", "\\.", -1)
	wdRegex = "^" + strings.Replace(wdRegex, "*", "[^.]*", -1) + "$"
	return regexp.MatchString(wdRegex, route)
}

type RouteExt struct {
	route ccv2.Route
	hostname string
}

func getMatchingRoutes(cliConnection plugin.CliConnection, hostname string) (matchingRoutes []RouteExt, err error) {
	matchingRoutes = make([]RouteExt, 0)
	domain, domainFound, err := getDomain(cliConnection, hostname)
	if err != nil {
		return
	}
	if !domainFound {
		err = errors.New("could not find matching domain")
		return
	}

	routes, err := getDomainRoutes(cliConnection, domain)
	if err != nil {
		return
	}

	for _, route := range routes {
		routeHostname := domain.Name
		if route.Host != "" {
			routeHostname = route.Host + "." + routeHostname
		}
		matched, _ := matchWildcard(hostname, routeHostname)
		if matched {
			matchingRoutes = append(matchingRoutes, RouteExt{route: route, hostname: routeHostname})
		}
	}

	return
}

type AppExt struct {
	route RouteExt
	app App
}

func getApps(cliConnection plugin.CliConnection, hostname string) (appsExt []AppExt, err error) {
	routes, err := getMatchingRoutes(cliConnection, hostname)
	if err != nil {
		return
	}
	if len(routes) == 0 {
		err = errors.New("route not found")
		return
	}

	appsExt = make([]AppExt, 0)

	for _, route := range routes {
		uri := "/v2/routes/" + route.route.GUID + "/apps"
		appCount := 0
		// paginate
		for uri != "" {
			var body string
			body, err = apiCall(cliConnection, uri)
			if err != nil {
				return
			}

			var data AppsResponse
			err = json.Unmarshal([]byte(body), &data)
			if err != nil {
				return
			}

			for _, res := range data.Resources {
				ra := AppExt{route: route, app: res}
				appsExt = append(appsExt, ra)
				appCount++
			}
			uri = data.NextUrl
		}
		if appCount == 0 {
			appsExt = append(appsExt, AppExt{route: route, app: App{}})
		}
	}

	return
}
