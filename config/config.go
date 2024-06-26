package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SlackToken    string
	SlackAppToken string
	SlackChannel  string
	AdminUserID   string
	BaseURL       string
}

var config *Config

func GetConfig() *Config {
	if config == nil {
		config = NewConfig()
	}

	return config
}

func NewConfig() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	slackToken, ok := os.LookupEnv("SLACK_TOKEN")
	if !ok {
		log.Fatal("SLACK_TOKEN is not set")
	}
	slackAppToken, ok := os.LookupEnv("SLACK_APP_TOKEN")
	if !ok {
		log.Fatal("SLACK_APP_TOKEN is not set")
	}
	slackChannel, ok := os.LookupEnv("SLACK_CHANNEL")
	if !ok {
		log.Fatal("SLACK_CHANNEL is not set")
	}
	adminUserID, ok := os.LookupEnv("ADMIN_USER_ID")
	if !ok {
		log.Fatal("ADMIN_USER_ID is not set")
	}

	baseURL, ok := os.LookupEnv("BASE_URL")
	if !ok {
		log.Fatal("BASE_URL is not set")
	}

	return &Config{
		SlackToken:    slackToken,
		SlackAppToken: slackAppToken,
		SlackChannel:  slackChannel,
		AdminUserID:   adminUserID,
		BaseURL:       baseURL,
	}
}
