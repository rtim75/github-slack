package subscription

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/rtim75/githubToSlack/update"

	"gopkg.in/yaml.v2"
)

type SubscriptionsConfig struct {
	// User         string   `yaml:"user"`
	Channel      string   `yaml:"channel"`
	Repositories []string `yaml:"repositories"`
	Interested   []string `yaml:"interested"`
}

type subscription struct {
	channel     string
	repository  string
	interested  []string
	lastUpdated time.Time
}

func initConfig(filePath string) ([]subscription, error) {

	var (
		subConfigs []SubscriptionsConfig
		subs       []subscription
	)

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	yaml.Unmarshal(file, &subConfigs)

	for _, subConfig := range subConfigs {
		for _, repo := range subConfig.Repositories {
			if !strings.HasPrefix(repo, "https://") {
				repo = "https://" + repo
			}

			subs = append(subs, subscription{
				channel:    subConfig.Channel,
				repository: repo,
				interested: subConfig.Interested,
			})
		}
	}
	return subs, nil
}

func Watch(filepath string) {
	fp := gofeed.NewParser()

	subs, err := initConfig(filepath)
	if err != nil {
		log.Fatalf("Failed to initialize a config: %v", err)
	}

	for i := range subs {
		feed, err := fp.ParseURL(subs[i].repository + "/releases.atom")
		if err != nil {
			fmt.Printf("Failed to fetch the feed: %v", err)
		}
		subs[i].lastUpdated = *feed.Items[0].UpdatedParsed
	}

	ticker := time.NewTicker(2 * time.Minute)

	for {
		select {
		case <-ticker.C:
			for i := range subs {
				log.Printf("Checking update for %v", subs[i].repository)

				feed, _ := fp.ParseURL(subs[i].repository + "/releases.atom")

				if feed.Items[0].UpdatedParsed.After(subs[i].lastUpdated) {
					subs[i].lastUpdated = *feed.Items[0].UpdatedParsed
					update.Post(subs[i].channel, subs[i].repository, feed.Items[0].Title, subs[i].lastUpdated)
					log.Printf("New release of %v: %v at %v\n", subs[i].repository, feed.Items[0].Title, feed.Items[0].UpdatedParsed)
				}
			}
		}
	}
}
