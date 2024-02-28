package proxy

import (
	"io"
	"log"
	"net/http"
	"proxy/internal/pkg/api"
)

type Proxy struct {
	uc api.Usecase
}

func NewProxy(uc api.Usecase) *Proxy {
	return &Proxy{
		uc: uc,
	}
}

func (p *Proxy) HTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Scheme = "http"
	copyRequest := *r
	r.RequestURI = ""
	r.Header.Del("Proxy-Connection")

	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	for _, cookie := range resp.Cookies() {
		http.SetCookie(w, cookie)
	}

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Println("Error copying response body:", err)
	}

	_, err = p.uc.SaveRequestResponse(&copyRequest, resp)
	if err != nil {
		log.Println("Error saving request_response:", err)
	}
}
