package main

import (
	"log"
	"net/http"
	"os"

	"github.com/asim9115/containerix/router"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("starting containerix on port", port)
	log.Fatal(http.ListenAndServe(":"+port, router.NewRouter()))
}