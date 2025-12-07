package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	myapi "FinalProject/pkg/api"
	mydb "FinalProject/pkg/db"
)

func main() {
	webDir := "./web"

	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	if err := mydb.Init(dbFile); err != nil {
		log.Fatal("DB init error: ", err)
	}
	defer mydb.Close()
	myapi.Init()

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	fmt.Println("LOCAL:", time.Now())
	fmt.Println("UTC:  ", time.Now().UTC())

	log.Printf("Server started on port %s", port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}

}
