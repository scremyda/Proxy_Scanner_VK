package models

type Response struct {
	ContentLength int64     `json:"contentLength"`
	RequestId     int       `json:"requestID"`
	Code          int       `json:"code"`
	Message       string    `json:"message"`
	Cookies       []Cookies `json:"cookies"`
	Headers       string    `json:"headers"`
	Body          string    `json:"body"`
}
