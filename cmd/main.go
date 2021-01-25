package main

import (
	"fmt"
	"linetarot/function"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello")
	})
	r.HandleFunc("/webhook", function.RandomCard).Methods("POST")

	http.ListenAndServe(":8080", r)
}
