package handler

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/nfnt/resize"
	"github.com/slack-go/slack"
	"github.com/xxarupakaxx/emoji_add_notification_bot/config"
)

func HandleReaction(evt *slack.ReactionAddedEvent, client *slack.Client) error {
	if evt.Reaction != "done" {
		log.Println("reaction is not done")
		return fmt.Errorf("reaction is not done")
	}

	if config.GetConfig().AdminUserID == "" {
		log.Println("admin user id is not set")
		return fmt.Errorf("admin user id is not set")
	}

	msgInfo := slack.NewRefToMessage(evt.Item.Channel, evt.Item.Timestamp)
	history, err := client.GetConversationHistory(&slack.GetConversationHistoryParameters{
		ChannelID: evt.Item.Channel,
		Latest:    evt.Item.Timestamp,
		Limit:     1,
		Inclusive: true,
	})
	if err != nil {
		log.Println("failed to get conversation history", err)
		return fmt.Errorf("failed to get conversation history: %w", err)
	}

	if len(history.Messages) == 0 || len(history.Messages[0].Files) == 0 {
		log.Println("no message found OR no file found")
		return nil
	}

	file := history.Messages[0].Files[0]

	resp, err := http.Get(file.URLPrivateDownload)
	if err != nil {
		log.Println("failed to download file", err)
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Println("failed to decode image", err)
		return fmt.Errorf("failed to decode image: %w", err)
	}

	resizedImage := resize.Resize(128, 128, img, resize.Lanczos3)

	emojiName := fmt.Sprintf("%s", file.Name[:len(file.Name)-len(file.Filetype)-1])
	emojiName = strings.ReplaceAll(emojiName, " ", "")

	tempFile, err := os.CreateTemp("", "emoji_*.png")
	if err != nil {
		log.Println("failed to create temp file", err)
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	defer os.Remove(tempFile.Name())

	png.Encode(tempFile, resizedImage)
	tempFile.Close()
	if err := addEmoji(emojiName, tempFile.Name()); err != nil {
		log.Println("failed to add emoji", err)
		return fmt.Errorf("failed to add emoji: %w", err)
	}

	return nil
}

func addEmoji(name, path string) error {
	return nil
}
