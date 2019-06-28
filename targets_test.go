package main

import (
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTargets(t *testing.T) {
	targets := NewTargets()
	assert.Equal(t, 0, len(targets.Orgs), "Orgs map is empty")
	assert.Equal(t, 0, len(targets.Spaces), "Spaces map is empty")
}

func TestTargets_getSpaceByGuid(t *testing.T) {
	targets := NewTargets()
	fakeCliConnection := &pluginfakes.FakeCliConnection{}
	apiCallCount := 0
	spaceName := "BEST-REGARDS-FROM-LUBLIN-AT-POLAND-SPACE"
	fakeCliConnection.CliCommandWithoutTerminalOutputStub = func(args ...string) ([]string, error) {
		apiCallCount++
		return []string{`{
			"metadata": {
				"guid": "bc8d3381-390d-4bd7-8c71-25309900a2e3",
				"url": "/v2/spaces/bc8d3381-390d-4bd7-8c71-25309900a2e3",
				"created_at": "2016-06-08T16:41:40Z",
				"updated_at": "2016-06-08T16:41:26Z"
			},
			"entity": {
				"name": "`+spaceName+`",
				"organization_guid": "6e1ca5aa-55f1-4110-a97f-1f3473e771b9",
				"space_quota_definition_guid": null,
				"allow_ssh": true,
				"organization_url": "/v2/organizations/6e1ca5aa-55f1-4110-a97f-1f3473e771b9",
				"developers_url": "/v2/spaces/bc8d3381-390d-4bd7-8c71-25309900a2e3/developers",
				"managers_url": "/v2/spaces/bc8d3381-390d-4bd7-8c71-25309900a2e3/managers",
				"auditors_url": "/v2/spaces/bc8d3381-390d-4bd7-8c71-25309900a2e3/auditors",
				"apps_url": "/v2/spaces/bc8d3381-390d-4bd7-8c71-25309900a2e3/apps",
				"routes_url": "/v2/spaces/bc8d3381-390d-4bd7-8c71-25309900a2e3/routes",
				"domains_url": "/v2/spaces/bc8d3381-390d-4bd7-8c71-25309900a2e3/domains",
				"service_instances_url": "/v2/spaces/bc8d3381-390d-4bd7-8c71-25309900a2e3/service_instances",
				"app_events_url": "/v2/spaces/bc8d3381-390d-4bd7-8c71-25309900a2e3/app_events",
				"events_url": "/v2/spaces/bc8d3381-390d-4bd7-8c71-25309900a2e3/events",
				"security_groups_url": "/v2/spaces/bc8d3381-390d-4bd7-8c71-25309900a2e3/security_groups",
				"staging_security_groups_url": "/v2/spaces/bc8d3381-390d-4bd7-8c71-25309900a2e3/staging_security_groups"
			}
		}`}, nil
	}
	assert.Equal(t, 0, apiCallCount, "API hasn't been called")
	space, err := targets.GetSpaceByGuid(fakeCliConnection, "guid-1")
	assert.Equal(t, nil, err, "Without error")
	assert.Equal(t, spaceName, space.Entity.Name, "Space name as expected")
	assert.Equal(t, 1, apiCallCount, "API has been called")

	space, err = targets.GetSpaceByGuid(fakeCliConnection, "guid-1")
	assert.Equal(t, nil, err, "Without error")
	assert.Equal(t, spaceName, space.Entity.Name, "Space name as expected")
	assert.Equal(t, 1, apiCallCount, "API hasn't been called one more time, cache is working")

	space, err = targets.GetSpaceByGuid(fakeCliConnection, "guid-2")
	assert.Equal(t, nil, err, "Without error")
	assert.Equal(t, spaceName, space.Entity.Name, "Space name as expected")
	assert.Equal(t, 2, apiCallCount, "API has been called for the new guid")
}

func TestTargets_getOrgByGuid(t *testing.T) {
	targets := NewTargets()
	fakeCliConnection := &pluginfakes.FakeCliConnection{}
	apiCallCount := 0
	orgName := "BEST-REGARDS-FROM-LUBLIN-AT-POLAND-ORG"
	fakeCliConnection.CliCommandWithoutTerminalOutputStub = func(args ...string) ([]string, error) {
		apiCallCount++
		return []string{`{
			"metadata": {
				"guid": "1c0e6074-777f-450e-9abc-c42f39d9b75b",
				"url": "/v2/organizations/1c0e6074-777f-450e-9abc-c42f39d9b75b",
				"created_at": "2016-06-08T16:41:33Z",
				"updated_at": "2016-06-08T16:41:26Z"
			},
			"entity": {
				"name": "`+ orgName +`",
				"billing_enabled": false,
				"quota_definition_guid": "769e777f-92b6-4ba0-9e48-5f77e6293670",
				"status": "active",
				"quota_definition_url": "/v2/quota_definitions/769e777f-92b6-4ba0-9e48-5f77e6293670",
				"spaces_url": "/v2/organizations/1c0e6074-777f-450e-9abc-c42f39d9b75b/spaces",
				"domains_url": "/v2/organizations/1c0e6074-777f-450e-9abc-c42f39d9b75b/domains",
				"private_domains_url": "/v2/organizations/1c0e6074-777f-450e-9abc-c42f39d9b75b/private_domains",
				"users_url": "/v2/organizations/1c0e6074-777f-450e-9abc-c42f39d9b75b/users",
				"managers_url": "/v2/organizations/1c0e6074-777f-450e-9abc-c42f39d9b75b/managers",
				"billing_managers_url": "/v2/organizations/1c0e6074-777f-450e-9abc-c42f39d9b75b/billing_managers",
				"auditors_url": "/v2/organizations/1c0e6074-777f-450e-9abc-c42f39d9b75b/auditors",
				"app_events_url": "/v2/organizations/1c0e6074-777f-450e-9abc-c42f39d9b75b/app_events",
				"space_quota_definitions_url": "/v2/organizations/1c0e6074-777f-450e-9abc-c42f39d9b75b/space_quota_definitions"
			}
		}`}, nil
	}
	assert.Equal(t, 0, apiCallCount, "API hasn't been called")
	space, err := targets.GetOrgByGuid(fakeCliConnection, "guid-1")
	assert.Equal(t, nil, err, "Without error")
	assert.Equal(t, orgName, space.Entity.Name, "Org name as expected")
	assert.Equal(t, 1, apiCallCount, "API has been called")

	space, err = targets.GetOrgByGuid(fakeCliConnection, "guid-1")
	assert.Equal(t, nil, err, "Without error")
	assert.Equal(t, orgName, space.Entity.Name, "Org name as expected")
	assert.Equal(t, 1, apiCallCount, "API hasn't been called one more time, cache is working")

	space, err = targets.GetOrgByGuid(fakeCliConnection, "guid-2")
	assert.Equal(t, nil, err, "Without error")
	assert.Equal(t, orgName, space.Entity.Name, "Org name as expected")
	assert.Equal(t, 2, apiCallCount, "API has been called for the new guid")
}
