package main

import (
	"log"
	"net/http"
)

func main() {
	filePathRoot := "."
	port := "8080"
	serveMux := http.NewServeMux()
	server := http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}
	serveMux.Handle("/", http.FileServer(http.Dir(filePathRoot)))
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Error:", err)
	}
}
