package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/xxarupakaxx/emoji_add_notification_bot/config"
	"github.com/xxarupakaxx/emoji_add_notification_bot/handler"
)

func main() {
	conf := config.NewConfig()
	client := slack.New(conf.SlackToken, slack.OptionAppLevelToken(conf.SlackAppToken))

	socketClient := socketmode.New(client, socketmode.OptionDebug(true), socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	os.MkdirAll("tmp", os.ModePerm)

	go handler.StartImageServer()

	go func(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {
		for {
			select {
			case <-ctx.Done():
				log.Println("context is done")
				return
			case e := <-socketClient.Events:
				switch e.Type {
				case socketmode.EventTypeEventsAPI:
					eventsAPIEvent, ok := e.Data.(slackevents.EventsAPIEvent)
					if !ok {
						log.Printf("Ignored %+v\n", e)
						continue
					}

					socketClient.Ack(*e.Request)

					switch eventsAPIEvent.Type {
					case slackevents.CallbackEvent:
						innnerEvent := eventsAPIEvent.InnerEvent
						switch ev := innnerEvent.Data.(type) {
						case *slackevents.ReactionAddedEvent:
							if err := handler.HandleReaction(ev, client); err != nil {
								log.Println("failed to handle reaction", err)
							}
						case *slackevents.EmojiChangedEvent:
							fmt.Println("emoji changed")
							if err := handler.HandleNoticeEmoji(ev, client); err != nil {
								log.Println("failed to handle emoji add", err)
							}
						}
					}
				}
			}
		}
	}(ctx, client, socketClient)

	socketClient.Run()
}
