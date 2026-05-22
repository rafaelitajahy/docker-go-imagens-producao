package main

import (
	"log"
	"net/http"
)

func main() {
	server := &http.Server{
		Addr: ":8080",
	}

	log.Println("API rodando na porta 8080")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
