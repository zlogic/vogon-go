package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/handlers"
	"github.com/zlogic/vogon-go/data"
	"github.com/zlogic/vogon-go/server"
)

func serve(db data.DBService) {
	// Create the router and webserver
	services, err := server.CreateServices(db)
	if err != nil {
		log.WithError(err).Error("Error while creating services")
		return
	}
	router, err := server.CreateRouter(services)
	if err != nil {
		log.WithError(err).Error("Error while creating services")
		return
	}

	errs := make(chan error, 2)
	go func() {
		errs <- http.ListenAndServe(":8080", handlers.CompressHandler(router))
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	<-errs
}

func main() {
	// Init data layer
	db, err := data.Open(data.DefaultOptions())
	defer db.Close()
	if err != nil {
		db.Close()
		log.Fatalf("Failed to open data store %v", err)
	}

	if len(os.Args) < 2 || os.Args[1] == "serve" {
		serve(db)
	} else {
		switch directive := os.Args[1]; directive {
		default:
			log.Fatalf("Unrecognized directive %v", directive)
		}
	}
}
