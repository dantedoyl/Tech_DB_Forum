package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx"
	"log"
	"net/http"

	handler "github.com/dantedoyl/Tech_DB_Forum/internal/forum/delivery/http"
	repo "github.com/dantedoyl/Tech_DB_Forum/internal/forum/repository/postgres"
)

func main() {
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()

	dbConnPool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     pgx.ConnConfig{
			Port:                 5432,
			Database:             "techdb",
			User:                 "dantedoyl",
			Password:             "mypassword",
			PreferSimpleProtocol: true,
		},
		MaxConnections: 100,
		AfterConnect:   nil,
		AcquireTimeout: 0,
	})
	if err != nil {
		log.Fatal(err)
	}

	forumRepo := repo.NewForumRepository(dbConnPool)
	handler.NewForumHandler(api, forumRepo)

	err = http.ListenAndServe(":5000", api)
	if err != nil {
		log.Fatal("can't start server")
	}
	fmt.Println("Starting server on port :5000")
}
