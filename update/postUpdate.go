package update

import (
	"fmt"
	"log"
	"time"

	"github.com/slack-go/slack"
)

const update_tpl = "%v was tagged with %v at %v"

func Post(channel string, repo string, tag string, timestamp time.Time) {
	api := slack.New("")

	update := fmt.Sprintf(update_tpl, repo, tag, timestamp)

	_, _, err := api.PostMessage(
		channel,
		slack.MsgOptionAsUser(true),
		slack.MsgOptionText(update, false))

	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Update successfully sent to channel %s: %v\n", channel, update)
}
