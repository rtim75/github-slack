package main

import "time"

type repositoryLatestRelease struct {
	name      string
	latestTag string
	released  time.Time
}
