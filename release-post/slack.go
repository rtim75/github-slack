package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

type repository struct {
	Name  string    `json:"name"`
	Tag   string    `json:"tag"`
	TagTS time.Time `json:"tag_ts"`
}

func sendSlackUpdate(event string, token string) {
	const updateTemplate = "%v has been tagged with %v. Checkout the changelog at %v."

	repo := &repository{}
	json.Unmarshal([]byte(event), repo)

	projectName := strings.TrimLeft(repo.Name, "github.com")
	releasePage := fmt.Sprintf("%v/releases/tag/%v", repo.Name, repo.Tag)
	update := fmt.Sprintf(updateTemplate, projectName, repo.Tag, releasePage)

	api := slack.New(token)
}
