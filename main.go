package main

import (

	"log"
	"os"
)

var err error

func main() {

	r := registerRoutes()

	port := os.Getenv("PORT")

	if port == "" {
		port = "5000"
	}

	log.Printf("Listening on port %s\n", port)
	r.Run(":" + port)
}
