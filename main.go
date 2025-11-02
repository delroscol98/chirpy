package main

import (
	"log"
	"net/http"
)

func main() {
	filePathRoot := "."
	port := "8080"
	serveMux := http.NewServeMux()
	serveMux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir(filePathRoot))))
	serveMux.HandleFunc("/healthz", handler)
	server := &http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Error:", err)
	}
}

func handler(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}
