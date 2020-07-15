package main

import (
	"github.com/gdsoumya/better_ci/ci"
	"github.com/gorilla/mux"
	"log"
	"net/http"
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