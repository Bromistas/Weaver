package common

import "time"

type Message struct {
	ID       string
	Body     string
	Visible  bool
	Received time.Time
}
