package websocket

import "time"

const (
	// Ping interval to check client connection
	pingPeriod = time.Duration(15) * time.Second

	// chanel of client messages size
	outChannelSize = 500
)

type WSEvent struct {
	UserID  string `json:"user_id"`
	Matched bool   `json:"matched"`
}

var Events = make(chan WSEvent)
