package main

import (
	"github.com/rtim75/githubToSlack/subscription"
)

func main() {
	// users, err := api.GetUsers()

	// if err != nil {
	// 	log.Fatalf("Failed to get users: %v", err)
	// }
	// for _, user := range users {
	// 	fmt.Println(user.)
	// }

	// user, err := api.GetUserInfo("")
	// user, err := api.GetUserByEmail("")

	// if err != nil {
	// 	log.Fatalf("Failed to get user info: %v", err)
	// }

	// channel, _, _, err := api.OpenConversation(&slack.OpenConversationParameters{
	// 	Users: []string{user.ID},
	// })

	// if err != nil {
	// 	log.Fatalf("Failed to open conversation: %v", err)
	// }

	// api.PostMessage(channel.ID, slack.MsgOptionText("hello Ruslan", false))
	// subscription.Watch()
	subscription.Watch("subscription_list.yml")
}
