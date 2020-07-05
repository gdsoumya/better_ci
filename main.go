package main

import (
	"log"
	"net/http"

	"github.com/gdsoumya/better_ci/ci"
	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	c, err := ci.Init()
	if err != nil {
		return
	}
	router.HandleFunc("/payload", c.WebHook)
	router.HandleFunc("/", c.Ping)
	log.Print("Server Started at :8080")
	http.ListenAndServe(":8080", router)
}
