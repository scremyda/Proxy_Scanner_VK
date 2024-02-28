package api

import (
	"net/http"
	"proxy/internal/models"
)

type Usecase interface {
	SaveRequestResponse(r *http.Request, response *http.Response) (models.Response, error)
	AllRequests() ([]models.Request, error)
	GetRequest(id int) (models.Request, error)
	RepeatRequest(id int) (models.Response, error)
	Scan(id int) (*models.CommandInjection, error)
}

type Repo interface {
	SaveRequestResponse(request models.Request, response models.Response) error
	GetRequest(requestId int) (models.Request, error)
	AllRequests() ([]models.Request, error)
}
