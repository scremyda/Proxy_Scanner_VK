package proxy

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"proxy/internal/pkg/utils"
	"syscall"
)

func (p *Proxy) HTTPS(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		log.Fatal("http server doesn't support hijacking connection")
	}

	clientConn, _, err := hj.Hijack()
	if err != nil {
		log.Fatal("http hijacking failed")
	}
	defer clientConn.Close()

	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		log.Fatal("error splitting host/port:", err)
	}

	caCert, caKey, err := utils.LoadX509KeyPair(host)
	if err != nil {
		log.Fatal("LoadX509KeyPair failed")
	}

	pemCert, pemKey := utils.CreateCert([]string{host}, caCert, caKey, 240)
	tlsCert, err := tls.X509KeyPair(pemCert, pemKey)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")); err != nil {
		log.Fatal("error writing status to client:", err)
	}

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
		MinVersion:               tls.VersionTLS13,
		Certificates:             []tls.Certificate{tlsCert},
	}

	tlsConn := tls.Server(clientConn, tlsConfig)
	defer tlsConn.Close()

	connReader := bufio.NewReader(tlsConn)

	request, err := http.ReadRequest(connReader)
	if err == io.EOF {
		log.Fatal(request, err)
	} else if errors.Is(err, syscall.ECONNRESET) {
		log.Print("This is connection reset by peer error")
		log.Fatal(err)
	} else if err != nil {
		log.Fatal(err)
	}

	if b, err := httputil.DumpRequest(request, false); err == nil {
		log.Printf("incoming request:\n%s\n", string(b))
	}

	copyReq := *request

	targetUrl := utils.AddrToUrl(request.Host)
	targetUrl.Path = request.URL.Path
	targetUrl.RawQuery = request.URL.RawQuery
	request.URL = targetUrl

	request.RequestURI = ""

	client := http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		log.Println("error sending request to target:", err)
		log.Fatal(request, err)
	}
	if b, err := httputil.DumpResponse(resp, true); err == nil {
		log.Printf("target response:\n%s\n", string(b))
	}
	defer resp.Body.Close()

	//copyResp := *resp

	if err := resp.Write(tlsConn); err != nil {
		log.Println("error writing response back:", err)
	}

	copyReq.URL.Scheme = "https"
	_, err = p.uc.SaveRequestResponse(&copyReq, resp)
	if err != nil {
		log.Println("Error saving request_response:", err)
	}

}
