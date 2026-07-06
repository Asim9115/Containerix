package main

import (
	"log"
	"net/http"
	"os"

	"github.com/asim9115/containerix/internal/state"
	"github.com/asim9115/containerix/router"
)

func main() {
	// Initialize global state: sandbox cgroup + port manager
	if err := state.Init("containerix", 2, "3221225472"); err != nil {
		log.Fatal(err)
	}
	log.Println("Sandbox and port manager ready")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("starting containerix on port", port)
	log.Fatal(http.ListenAndServe(":"+port, router.NewRouter()))
}