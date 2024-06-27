package handler

import (
	"fmt"
	"log"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/xxarupakaxx/emoji_add_notification_bot/config"
)

func HandleNoticeEmoji(evt *slackevents.EmojiChangedEvent, client *slack.Client) error {
	if evt.Subtype == "add" {
		// 新しい絵文字が追加された場合
		_, _, err := client.PostMessage(config.GetConfig().SlackChannel,
			slack.MsgOptionText(fmt.Sprintf("新しい絵文字 :%s: `%s` が追加されたぱか :clap-nya:", evt.Name, evt.Name), false))
		if err != nil {
			log.Println("error posting message", err)
			return fmt.Errorf("error posting message: %v", err)
		}
	}

	return nil
}
