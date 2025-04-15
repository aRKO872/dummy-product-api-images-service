package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/product-img-api-microservice/files"
	"github.com/product-img-api-microservice/handlers"

	gohandler "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	var basePath = "./imagestore"

	sm := mux.NewRouter()

	l := hclog.New(
		&hclog.LoggerOptions{
			Name:  "product-images",
		},
	)

	// create a logger for the server from the default logger
	sl := l.StandardLogger(&hclog.StandardLoggerOptions{InferLevels: true})

	// create the storage class, use local storage
	// max filesize 5MB
	stor, err := files.NewLocal(basePath, 1024*1000*5)
	if err != nil {
		l.Error("Unable to create storage", "error", err)
		os.Exit(1)
	}

	// create the handlers
	fh := handlers.NewFiles(stor, l)
	mw := handlers.GZipHandler{}

	ph := sm.Methods(http.MethodPost).Subrouter()
	ph.HandleFunc("/images/{id:[0-9]+}/{filename}", fh.UploadRest)
	ph.HandleFunc("/images", fh.UploadMultiPart)

	gh := sm.Methods(http.MethodGet).Subrouter()
	gh.Handle("/images/{id:[0-9]+}/{filename}", http.StripPrefix("/images", http.FileServer(http.Dir(basePath))))
	gh.Use(mw.GzipMiddleware)

	// CORS
	// We can add multiple referrers which can call this. Just adding one for the time being
	ch := gohandler.CORS(gohandler.AllowedOrigins([]string{"http://localhost:8000", "http://localhost:3000"}))

	// To allow everyone to access : 
	// ch := gohandler.CORS(gohandler.AllowedOrigins([]string{"*"}))

	// create a new server
	s := http.Server{
		Addr:         ":8080",      // configure the bind address
		Handler:      ch(sm),                // set the default handler
		ErrorLog:     sl,                // the logger for the server
		ReadTimeout:  5 * time.Second,   // max time to read request from the client
		WriteTimeout: 10 * time.Second,  // max time to write response to the client
		IdleTimeout:  120 * time.Second, // max time for connections using TCP Keep-Alive
	}

	go func () {
		err := s.ListenAndServe()
		if err != nil {
			l.Error(err.Error())
			os.Exit(1)
		}
	}()

	// Process for graceful shutdown

	sigChan := make(chan os.Signal, 2)

	signal.Notify(sigChan, os.Interrupt)

	sig := <-sigChan

	l.Info("Graceful Shutdown, due to :", sig)

	tc, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(tc)
}