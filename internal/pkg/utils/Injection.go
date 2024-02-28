package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"proxy/internal/models"
	"strings"
)

func AddrToUrl(addr string) *url.URL {
	if !strings.HasPrefix(addr, "https") {
		addr = "https://" + addr
	}
	u, err := url.Parse(addr)
	if err != nil {
		log.Fatal(err)
	}
	return u
}

func CheckCommandInjection(req *http.Request) (*models.CommandInjection, error) {
	scanInjections := &models.CommandInjection{}
	var err error
	scanInjections.CommandInjectionInCookie, err = checkCookieInjection(req)
	if scanInjections.CommandInjectionInCookie == "" {
		scanInjections.CommandInjectionInCookie = "No vulnerabilities"
	}
	if err != nil {
		return nil, err
	}

	scanInjections.CommandInjectionInGetParams, err = checkGETInjection(req)
	if scanInjections.CommandInjectionInGetParams == "" {
		scanInjections.CommandInjectionInGetParams = "No vulnerabilities"
	}
	if err != nil {
		return nil, err
	}

	if req.Method == "POST" {
		scanInjections.CommandInjectionInPostParams, err = checkPOSTInjection(req)
		if err != nil {
			return nil, err
		}
	}

	if scanInjections.CommandInjectionInPostParams == "" {
		scanInjections.CommandInjectionInPostParams = "No vulnerabilities"
	}

	scanInjections.CommandInjectionInHeaders, err = checkHeaderInjection(req)
	if scanInjections.CommandInjectionInHeaders == "" {
		scanInjections.CommandInjectionInHeaders = "No vulnerabilities"
	}
	if err != nil {
		return nil, err
	}

	return scanInjections, nil

}

func checkCookieInjection(req *http.Request) (string, error) {
	payloads := []string{";cat /etc/passwd;", "|cat /etc/passwd|", "`cat /etc/passwd`"}

	for _, payload := range payloads {
		req.AddCookie(&http.Cookie{Name: "cookie_name", Value: payload})

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Ошибка при отправке запроса: %v\n", err)
			return "", err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Ошибка при чтении тела ответа: %v\n", err)
			return "", err
		}

		// Проверка наличия "root:" в теле ответа
		if strings.Contains(string(body), "root:") {
			return fmt.Sprintf("Обнаружена уязвимость в Cookie: %s\n", payload), nil
		}
	}

	return "", nil
}

func checkGETInjection(req *http.Request) (string, error) {
	payloads := []string{";cat /etc/passwd;", "|cat /etc/passwd|", "`cat /etc/passwd`"}

	for _, payload := range payloads {
		req.URL.RawQuery = fmt.Sprintf("param=%s", payload)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Ошибка при отправке запроса: %v\n", err)
			return "", err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Ошибка при чтении тела ответа: %v\n", err)
			return "", err
		}

		if strings.Contains(string(body), "root:") {
			return fmt.Sprintf("Обнаружена уязвимость в GET параметрах: %s\n", payload), nil
		}
	}

	return "", nil
}

func checkHeaderInjection(req *http.Request) (string, error) {
	payloads := []string{";cat /etc/passwd;", "|cat /etc/passwd|", "`cat /etc/passwd`"}

	for _, payload := range payloads {
		req.Header.Set("User-Agent", payload)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Ошибка при отправке запроса: %v\n", err)
			return "", err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Ошибка при чтении тела ответа: %v\n", err)
			return "", err
		}

		if strings.Contains(string(body), "root:") {
			return fmt.Sprintf("Обнаружена уязвимость в HTTP заголовках: %s\n", payload), nil
		}
	}

	return "", nil
}

func checkPOSTInjection(req *http.Request) (string, error) {
	payloads := []string{";cat /etc/passwd;", "|cat /etc/passwd|", "`cat /etc/passwd`"}

	for _, payload := range payloads {
		req.PostForm = map[string][]string{"param": {payload}}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Ошибка при отправке запроса: %v\n", err)
			return "", err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Ошибка при чтении тела ответа: %v\n", err)
			return "", err
		}

		if strings.Contains(string(body), "root:") {
			return fmt.Sprintf("Обнаружена уязвимость в POST параметрах: %s\n", payload), nil
		}
	}

	return "", nil
}

func ConvertModelToRequest(request models.Request) (*http.Request, error) {
	httpRequest, err := http.NewRequest(request.Method, request.Scheme+"://"+request.Host+request.Path, strings.NewReader(request.Body))
	if err != nil {
		return nil, err
	}

	// Set headers
	strToHeaders := func(headersString string) map[string]string {
		headers := make(map[string]string)

		lines := strings.Split(headersString, "\n")
		for _, line := range lines {
			if line != "" {
				parts := strings.SplitN(line, " ", 2)
				if len(parts) == 2 {
					key := parts[0]
					value := parts[1]
					headers[key] = value
				}
			}
		}

		return headers
	}

	for key, value := range strToHeaders(request.Headers) {
		httpRequest.Header.Set(key, value)
	}

	// Set cookies
	for _, cookie := range request.Cookies {
		httpRequest.AddCookie(&http.Cookie{
			Name:  cookie.Key,
			Value: cookie.Value,
		})
	}

	// Set GET parameters
	queryParams := make(url.Values)
	for _, value := range request.GetParams {
		queryParams.Add(value.Key, value.Value)
	}
	httpRequest.URL.RawQuery = queryParams.Encode()

	// Set POST parameters
	if request.Method == "POST" && len(request.PostParams) > 0 {
		formData := make(url.Values)
		for _, param := range request.PostParams {
			formData.Add(param.Key, param.Value)
		}
		httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		httpRequest.Body = io.NopCloser(strings.NewReader(formData.Encode()))
	}
	return httpRequest, nil
}
