package main

import (
	"gpt3.5/cfg"
	gptHttp "gpt3.5/http"
	"log"
	"net/http"
	"strconv"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/chat/do", http.HandlerFunc(gptHttp.Do))
	mux.Handle("/cfg/modifyKey", http.HandlerFunc(gptHttp.SwitchApikey))
	log.Printf("starting http-srv on port[%d]", cfg.Cfg.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(cfg.Cfg.Port), mux))
}
