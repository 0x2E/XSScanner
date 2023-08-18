package main

type Res struct {
	URL  string  `json:"url"`
	Vuln []*Vuln `json:"vuln"`
}

type Vuln struct {
	Title   string   `json:"title"`
	Param   param    `json:"param"`
	Req     string   `json:"req"`
	Resp    string   `json:"resp"`
	Attempt []string `json:"attempt"`
}
