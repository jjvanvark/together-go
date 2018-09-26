package main

import (
	"context"
	"flag"
	"log"
	"maus/together-go/database"
	"maus/together-go/routing"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	host   string
	dbHost string
	dbName string
)

func main() {

	// Flags

	flag.StringVar(&host, "host", ":9090", "server host address")
	flag.StringVar(&dbHost, "db-host", ":27017", "database host address")
	flag.StringVar(&dbName, "db-name", "togetherness", "database name")
	flag.Parse()

	// Database

	db, err := database.InitDatabase(dbHost, dbName)
	if err != nil {
		log.Fatal(err)
	}

	// Router

	router := routing.InitRouting("/v1", db)

	// Server

	server := &http.Server{Addr: host, Handler: router}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Stopping

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	log.Println("Start shutting down")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	server.Shutdown(ctx)

	routing.CloseRouting()

	db.Close()

	log.Println("End shutting down")

}
