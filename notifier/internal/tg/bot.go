package tg

import (
	"time"

	tb "gopkg.in/telebot.v4"
)

func NewBot(token string) (*tb.Bot, error) {
	return tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
}
