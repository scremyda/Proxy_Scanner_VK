package usecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"proxy/internal/models"
	"proxy/internal/pkg/api"
	"proxy/internal/pkg/utils"
	"strings"
)

type Usecase struct {
	repo api.Repo
}

func NewUsecase(repo api.Repo) *Usecase {
	return &Usecase{
		repo: repo,
	}
}

func (uc *Usecase) SaveRequestResponse(r *http.Request, response *http.Response) (models.Response, error) {
	request := &models.Request{
		Method: r.Method,
		Scheme: r.URL.Scheme,
		Host:   r.Host,
		Path:   r.URL.Path,
	}

	// Cookies
	for _, cookie := range r.Cookies() {
		request.Cookies = append(request.Cookies, models.Cookies{Key: cookie.Name, Value: cookie.Value})
	}

	// Headers
	headers, err := json.Marshal(r.Header)
	if err != nil {
		return models.Response{}, err
	}
	request.Headers = string(headers)

	// Body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return models.Response{}, err
	}
	request.Body = string(body)

	// GET parameters
	for key, values := range r.URL.Query() {
		for _, value := range values {
			request.GetParams = append(request.GetParams, models.Param{Key: key, Value: value})
		}
	}

	// POST parameters
	if err = r.ParseForm(); err != nil {
		return models.Response{}, err
	}
	for key, values := range r.PostForm {
		for _, value := range values {
			request.PostParams = append(request.PostParams, models.Param{Key: key, Value: value})
		}
	}

	responseModel := &models.Response{
		ContentLength: response.ContentLength,
		Code:          response.StatusCode,
		Message:       response.Status,
	}

	// Cookies
	for _, cookie := range response.Cookies() {
		responseModel.Cookies = append(responseModel.Cookies, models.Cookies{Key: cookie.Name, Value: cookie.Value})
	}

	// Headers
	responseHeaders, err := json.Marshal(response.Header)
	if err != nil {
		return models.Response{}, err
	}
	responseModel.Headers = string(responseHeaders)

	// Body
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return models.Response{}, err
	}
	responseModel.Body = string(responseBody)

	return *responseModel, uc.repo.SaveRequestResponse(*request, *responseModel)
}

func (uc *Usecase) AllRequests() ([]models.Request, error) {
	return uc.repo.AllRequests()
}

func (uc *Usecase) GetRequest(id int) (models.Request, error) {
	return uc.repo.GetRequest(id)
}

func (uc *Usecase) Scan(id int) (*models.CommandInjection, error) {
	request, err := uc.GetRequest(id)
	if err != nil {
		return nil, err
	}

	req, err := utils.ConvertModelToRequest(request)
	if err != nil {
		return nil, err
	}

	return utils.CheckCommandInjection(req)
}

func (uc *Usecase) RepeatRequest(id int) (models.Response, error) {
	request, err := uc.GetRequest(id)
	if err != nil {
		return models.Response{}, err
	}

	body := bytes.NewBufferString(request.Body)
	urlStr := request.Scheme + "://" + request.Host + request.Path
	for i, v := range request.GetParams {
		if i == 0 {
			urlStr += "?"
		}
		urlStr += v.Key + "=" + v.Value
	}

	req, err := http.NewRequest(request.Method, urlStr, body)

	if err != nil {
		fmt.Println(err)
		return models.Response{}, err
	}

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
		req.Header.Add(key, value)
	}

	copyReq := *req

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return models.Response{}, err
	}
	defer resp.Body.Close()

	res, err := uc.SaveRequestResponse(&copyReq, resp)
	if err != nil {
		log.Printf("Error save: %v", err)
	}

	return res, nil
}
