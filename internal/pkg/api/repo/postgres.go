package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgtype/pgxtype"
	"log"
	"proxy/internal/models"
)

const (
	insertRequestResponse = "INSERT INTO request_response(request, response) VALUES ($1, $2)"
	getRequestByID        = "SELECT request FROM request_response WHERE request_response_id = $1;"
	getAllRequests        = "SELECT request FROM request_response;"
)

type Repo struct {
	db pgxtype.Querier
}

func NewRepo(db pgxtype.Querier) *Repo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) SaveRequestResponse(request models.Request, response models.Response) error {
	requestStr, err := json.Marshal(request)
	if err != nil {
		err = fmt.Errorf("error happened in json.Marshal requestStr: %w", err)

		return err
	}

	responseStr, err := json.Marshal(response)
	if err != nil {
		err = fmt.Errorf("error happened in json.Marshal responseStr: %w", err)

		return err
	}

	_, err = r.db.Exec(
		context.Background(),
		insertRequestResponse,
		requestStr,
		responseStr,
	)
	if err != nil {
		err = fmt.Errorf("error happened in db.Exec SaveRequestResponse: %w", err)

		return err
	}

	return nil
}

func (r *Repo) GetRequest(requestId int) (models.Request, error) {
	requestStr := ""
	row := r.db.QueryRow(context.Background(), getRequestByID, requestId)
	err := row.Scan(
		&requestStr,
	)
	if err != nil {
		log.Println(err)
		return models.Request{}, err
	}

	var request models.Request
	err = json.Unmarshal([]byte(requestStr), &request)
	if err != nil {
		log.Println(err)
		return models.Request{}, err
	}
	return request, err

}

func (r *Repo) AllRequests() ([]models.Request, error) {
	requests := make([]models.Request, 0)
	rows, err := r.db.Query(context.Background(), getAllRequests)
	if err != nil {
		return requests, err
	}
	defer rows.Close()

	for rows.Next() {
		var request models.Request
		requestStr := ""
		err = rows.Scan(&requestStr)
		if err != nil {
			log.Println(err)
			return requests, err
		}

		err = json.Unmarshal([]byte(requestStr), &request)
		if err != nil {
			log.Println(err)
			return requests, err
		}

		requests = append(requests, request)
	}

	return requests, nil
}
