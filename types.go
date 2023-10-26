package main

// NodeInfoSchema : Refer to https://nodeinfo.diaspora.software/schema.html
type NodeInfoSchema struct {
	//Version  string `json:"version"`
	Software struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		//Repository *string `json:"repository,omitempty,omitempty"`
	} `json:"software"`
	//Protocols []string `json:"protocols"`
	//Services          any      `json:"services"`
	OpenRegistrations bool `json:"openRegistrations"`
	Usage             struct {
		Users struct {
			Total uint64 `json:"total"`
			//ActiveHalfyear *uint64 `json:"activeHalfyear,omitempty"`
			//ActiveMonth    *uint64 `json:"activeMonth,omitempty"`
		} `json:"users"`
		LocalPosts *uint64 `json:"localPosts,omitempty"`
		//LocalComments *uint64 `json:"localComments,omitempty"`
	} `json:"usage"`
	//Metadata any `json:"metadata"`
}

type NodeInfoList struct {
	Links []struct {
		Rel  string `json:"rel"` // http://nodeinfo.diaspora.software/ns/schema/2.1
		Href string `json:"href"`
	} `json:"links"`
}

type ErrorStatusWithCode struct {
	Offset int64 `json:"offset"`
	Code   int   `json:"code"`
}

type ResultErrRecord = map[string]int64
type ResultErrRecordWithCode = map[string]ErrorStatusWithCode
type ResultValidWithNodeInfo = map[string]NodeInfoSchema

type ResultFileFormat struct {
	CollectedAt int64 `json:"collected_at"`

	// Not working
	Unresolved     ResultErrRecord         `json:"unresolved"`
	NotFunctioning ResultErrRecord         `json:"not_functioning"`
	WrongCode      ResultErrRecordWithCode `json:"wrong_code"`

	// Working but failed to get node info
	MisformattedNodeInfoList   ResultErrRecord `json:"misformatted_nodeinfo_list"`
	NoAvailableNodeInfoSchema  ResultErrRecord `json:"no_available_nodeinfo_schema"`
	MisformattedNodeInfoSchema ResultErrRecord `json:"misformatted_nodeinfo_schema"`

	// Fully functional
	Valid ResultValidWithNodeInfo `json:"valid"`
}

// Temp structures
type domainWithErrorCode struct {
	Domain string
	Code   int
}
type domainWithValidNodeinfo struct {
	Domain   string
	NodeInfo NodeInfoSchema
}
