package handler

import (
	"log"
	"net/http"

	"github.com/xxarupakaxx/emoji_add_notification_bot/config"
)

func StartImageServer() {
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("tmp"))))
	log.Printf("Starting image server on %s\n", config.GetConfig().BaseURL+":1234")
	log.Fatal(http.ListenAndServe(":1234", nil))
}
