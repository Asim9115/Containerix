package main

import (
	"log"
	"net/http"
	"os"
	"github.com/asim9115/containerix/internal/sandbox"
	"github.com/asim9115/containerix/internal/state"
	"github.com/asim9115/containerix/router"
)



func main() {
		//intialize sandbox
	sb, err := sandbox.Init(
		"containerix",
		2,
		"3221225472",
	)
	if err != nil {
		log.Fatal(err)
	}
	state.SB.Sandbox = sb
	log.Println("Sandbox ready ")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("starting containerix on port", port)
	log.Fatal(http.ListenAndServe(":"+port, router.NewRouter()))
	

}