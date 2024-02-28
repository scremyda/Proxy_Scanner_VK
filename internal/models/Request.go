package models

type Request struct {
	Method     string    `json:"method"`
	Scheme     string    `json:"scheme"`
	Host       string    `json:"host"`
	Path       string    `json:"path"`
	Cookies    []Cookies `json:"cookies"`
	Headers    string    `json:"headers,omitempty"`
	Body       string    `json:"body,omitempty"`
	GetParams  []Param   `json:"getParams"`
	PostParams []Param   `json:"postParams"`
}

type Param struct {
	Key   string `json:"paramName"`
	Value string `json:"paramValue"`
}

type Cookies struct {
	Key   string `json:"cookieName"`
	Value string `json:"cookieValue"`
}
