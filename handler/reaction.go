package handler

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"strings"

	"github.com/nfnt/resize"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/xxarupakaxx/emoji_add_notification_bot/config"
)

func HandleReaction(evt *slackevents.ReactionAddedEvent, client *slack.Client) error {
	if evt.Reaction != "done" {
		log.Println("reaction is not done")
		return fmt.Errorf("reaction is not done")
	}

	if config.GetConfig().AdminUserID == "" {
		log.Println("admin user id is not set")
		return fmt.Errorf("admin user id is not set")
	}

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

	var fileContent bytes.Buffer

	if err := client.GetFile(file.URLPrivateDownload, &fileContent); err != nil {
		log.Println("failed to download file", err)
		return fmt.Errorf("failed to download file: %w", err)
	}

	fileType := http.DetectContentType(fileContent.Bytes())
	log.Printf("Detected file type: %s", fileType)

	var img image.Image
	switch fileType {
	case "image/jpeg":
		img, err = jpeg.Decode(&fileContent)
	case "image/png":
		img, err = png.Decode(&fileContent)
	case "image/gif":
		img, err = gif.Decode(&fileContent)
	default:
		log.Printf("Unsupported image format: %s", fileType)
		return fmt.Errorf("unsupported image format: %s", fileType)
	}

	resizedImage := resize.Resize(128, 128, img, resize.Lanczos3)

	emojiName := fmt.Sprintf("%s", file.Name[:len(file.Name)-len(file.Filetype)-1])
	emojiName = strings.ReplaceAll(emojiName, " ", "")

	var tempBuffer bytes.Buffer
	if err := png.Encode(&tempBuffer, resizedImage); err != nil {
		log.Printf("Error encoding resized image: %v", err)
		return fmt.Errorf("error encoding resized image: %w", err)
	}

	if err := addEmoji(client, emojiName, tempBuffer.Bytes()); err != nil {
		log.Println("failed to add emoji", err)
		return fmt.Errorf("failed to add emoji: %w", err)
	}

	_, _, err = client.PostMessage(evt.Item.Channel,
		slack.MsgOptionText(fmt.Sprintf("新しい絵文字 :%s: `%s` が追加されたぱか", emojiName, emojiName), false))
	if err != nil {
		log.Println("error posting message", err)
		return fmt.Errorf("error posting message: %v", err)
	}

	return nil
}

func addEmoji(client *slack.Client, emojiName string, image []byte) error {
	return nil
}
