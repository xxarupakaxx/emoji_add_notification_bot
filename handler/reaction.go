package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nfnt/resize"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/xxarupakaxx/emoji_add_notification_bot/config"
)

const SLACK_EMOJI_ADD_API = "https://slack.com/api/admin.emoji.add"

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

	imageURL, err := uploadImage(tempBuffer.Bytes())
	if err != nil {
		log.Printf("Error uploading image: %v", err)
		return fmt.Errorf("error uploading image: %w", err)
	}
	if err := addEmoji(emojiName, imageURL); err != nil {
		log.Println("failed to add emoji", err)
		return fmt.Errorf("failed to add emoji: %w", err)
	}

	_, _, err = client.PostMessage(evt.Item.Channel,
		slack.MsgOptionText(fmt.Sprintf("新しい絵文字 :%s: `%s` が追加されたぱか :clap-nya:", emojiName, emojiName), false))
	if err != nil {
		log.Println("error posting message", err)
		return fmt.Errorf("error posting message: %v", err)
	}

	return nil
}

func addEmoji(emojiName string, imageURL string) error {
	params := url.Values{}
	params.Add("token", config.GetConfig().SlackToken)
	params.Add("name", emojiName)
	params.Add("url", imageURL)

	req, err := http.NewRequest("GET", SLACK_EMOJI_ADD_API+"?"+params.Encode(), nil)
	if err != nil {
		log.Println("error creating request", err)
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+config.GetConfig().SlackToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("failed to do request", err)
		return fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	var slackResp struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &slackResp); err != nil {
		log.Println("error parsing response", err)
		return fmt.Errorf("error parsing response: %v", err)
	}

	if !slackResp.Ok {
		log.Println("Slack API error", slackResp.Error)
		return fmt.Errorf("Slack API error: %s", slackResp.Error)
	}

	return nil
}

func uploadImage(image []byte) (string, error) {
	filename := generateFilename()
	path := filepath.Join("tmp", filename)

	if err := os.WriteFile(path, image, 0o644); err != nil {
		log.Println("failed to write image", err)
		return "", fmt.Errorf("failed to write image: %w", err)
	}

	imageURL := fmt.Sprintf("%s/images/%s", config.GetConfig().BaseURL, filename)

	go func() {
		time.Sleep(1 * time.Minute)
		os.Remove(filepath.Join("tmp", filepath.Base(imageURL)))
	}()

	return imageURL, nil
}

func generateFilename() string {
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes) + ".png"
}
