package domain

import "time"

type Video struct {
	ID         string
	Title      string
	URL        string
	PlaylistID string
	Published  time.Time
}
