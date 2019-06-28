package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"encoding/json"
)

type Targets struct {
	Orgs   map[string]Org
	Spaces map[string]Space
}

func NewTargets() *Targets {
	t := new(Targets)
	t.Orgs = make(map[string]Org)
	t.Spaces = make(map[string]Space)
	return t
}

func (t Targets) GetSpaceByGuid(cliConnection plugin.CliConnection, guid string) (space Space, err error) {
	if t.Spaces[guid]==(Space{}) {
		var body string
		body, err = apiCall(cliConnection, "/v2/spaces/"+guid)
		if err != nil {
			return
		}
		err = json.Unmarshal([]byte(body), &space)
		if err != nil {
			return
		}
		t.Spaces[guid]=space
	} else {
		space = t.Spaces[guid]
	}
	return
}
func (t Targets) GetOrgByGuid(cliConnection plugin.CliConnection, guid string) (org Org, err error) {
	if t.Orgs[guid]==(Org{}) {
		var body string
		body, err = apiCall(cliConnection, "/v2/organizations/"+guid)
		if err != nil {
			return
		}
		err = json.Unmarshal([]byte(body), &org)
		if err != nil {
			return
		}
		t.Orgs[guid]=org
	} else {
		org = t.Orgs[guid]
	}
	return
}

func (t Targets) GetTargetBySpaceGuid(cliConnection plugin.CliConnection, spaceGuid string) (org Org, space Space, err error) {
	space, err = t.GetSpaceByGuid(cliConnection,spaceGuid)
	if err == nil {
		org, err = t.GetOrgByGuid(cliConnection,space.Entity.OrgGUID)
	}
	return
}

