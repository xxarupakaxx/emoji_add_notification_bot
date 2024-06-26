package handler

import (
	"log"

	"github.com/slack-go/slack"
	"github.com/xxarupakaxx/emoji_add_notification_bot/config"
)

func HandleReaction(evt *slack.ReactionAddedEvent, client *slack.Client) error {
	if evt.Reaction != "done" {
		log.Println("reaction is not done")
		return nil
	}

	if config.Config.AdminUserID == "" {
		log.Println("admin user id is not set")
		return nil
	}
}
