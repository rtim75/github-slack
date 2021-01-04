package main

import (
	"fmt"

	"github.com/mmcdole/gofeed"
)

func getLatestRelease(repository string) (*repositoryLatestRelease, error) {
	fp := gofeed.NewParser()

	feed, err := fp.ParseURL("https://" + repository + "/releases.atom")
	if err != nil {
		fmt.Printf("Failed to parse feed for %v: %v", repository, err)
		return nil, err
	}

	return &repositoryLatestRelease{
		name:      repository,
		latestTag: feed.Items[0].Title,
		released:  *feed.Items[0].UpdatedParsed,
	}, nil
}
