package main

import "time"

// NodeInfoSchema : Refer to https://nodeinfo.diaspora.software/schema.html
type NodeInfoSchema struct {
	//Version  string `json:"version"`
	Software struct {
		Name       string  `json:"name"`
		Version    string  `json:"version"`
		Repository *string `json:"repository,omitempty"`
	} `json:"software"`
	Protocols []string `json:"protocols"`
	//Services          any      `json:"services"`
	OpenRegistrations bool `json:"openRegistrations"`
	Usage             struct {
		Users struct {
			Total          uint64  `json:"total"`
			ActiveHalfyear *uint64 `json:"activeHalfyear"`
			ActiveMonth    *uint64 `json:"activeMonth"`
		} `json:"users"`
		LocalPosts    *uint64 `json:"localPosts"`
		LocalComments *uint64 `json:"localComments"`
	} `json:"usage"`
	//Metadata any `json:"metadata"`
}

type NodeInfoList struct {
	Links []struct {
		Rel  string `json:"rel"` // http://nodeinfo.diaspora.software/ns/schema/2.1
		Href string `json:"href"`
	} `json:"links"`
}

type DomainErrorStatus struct {
	Domain string    `json:"domain"`
	Since  time.Time `json:"since"`
}

type DomainErrorStatusWithCode struct {
	DomainErrorStatus

	Code int `json:"code"`
}

type DomainValidWithNodeinfo struct {
	Domain   string         `json:"domain"`
	NodeInfo NodeInfoSchema `json:"nodeinfo"`
}

type ResultFileFormat struct {
	CollectedAt time.Time `json:"collected_at"`

	// Not working
	Unresolved     []DomainErrorStatus         `json:"unresolved"`
	NotFunctioning []DomainErrorStatus         `json:"not_functioning"`
	WrongCode      []DomainErrorStatusWithCode `json:"wrong_code"`

	// Working but failed to get node info
	MisformattedNodeInfoList   []DomainErrorStatus `json:"misformatted_nodeinfo_list"`
	NoAvailableNodeInfoSchema  []DomainErrorStatus `json:"no_available_nodeinfo_schema"`
	MisformattedNodeInfoSchema []DomainErrorStatus `json:"misformatted_nodeinfo_schema"`

	// Fully functional
	Valid []DomainValidWithNodeinfo `json:"valid"`
}
