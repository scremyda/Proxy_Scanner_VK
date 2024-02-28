package models

type CommandInjection struct {
	CommandInjectionInCookie     string `json:"commandInjectionInCookie"`
	CommandInjectionInPostParams string `json:"commandInjectionInPostParams"`
	CommandInjectionInGetParams  string `json:"commandInjectionInGetParams"`
	CommandInjectionInHeaders    string `json:"commandInjectionInHeaders"`
}
