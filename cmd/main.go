package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/http"
	"os"
	"proxy/conf"
	"proxy/internal/pkg/api/delivery"
	"proxy/internal/pkg/api/repo"
	"proxy/internal/pkg/api/usecase"
	"proxy/internal/pkg/proxy"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() (err error) {
	cfg := conf.MustLoad()

	db, err := pgxpool.Connect(context.Background(), fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName))
	if err != nil {
		log.Println("fail open postgres", err)
		err = fmt.Errorf("error happened in sql.Open: %w", err)

		return err
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Println("fail ping postgres", err)
		err = fmt.Errorf("error happened in db.Ping: %w", err)

		return err
	}

	repo := repo.NewRepo(db)
	usecase := usecase.NewUsecase(repo)
	handler := delivery.NewHandler(usecase)

	r := mux.NewRouter()

	{
		r.HandleFunc("/requests", handler.AllRequests).
			Methods(http.MethodGet, http.MethodOptions)

		r.HandleFunc("/requests/{id:[0-9]+}", handler.GetRequestByID).
			Methods(http.MethodGet, http.MethodOptions)

		r.HandleFunc("/repeat/{id:[0-9]+}", handler.RepeatRequestByID).
			Methods(http.MethodGet, http.MethodOptions)

		r.HandleFunc("/scan/{id:[0-9]+}", handler.ScanByID).
			Methods(http.MethodGet, http.MethodOptions)
	}

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	http.Handle("/", r)

	apiServer := http.Server{
		Addr:    "127.0.0.1:8000",
		Handler: r,
	}
	go func() {
		if err := apiServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	pr := proxy.NewProxy(usecase)

	srv := http.Server{
		Addr: "127.0.0.1:8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				pr.HTTPS(w, r)
			} else {
				pr.HTTP(w, r)
			}
		}),
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

	return err
}
