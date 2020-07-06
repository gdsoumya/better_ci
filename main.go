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
	port := ":" + c.Port
	router.HandleFunc("/webhook", c.WebHook)
	router.HandleFunc("/", c.Ping)
	log.Print("Server Started at ", port)
	http.ListenAndServe(port, router)
}
