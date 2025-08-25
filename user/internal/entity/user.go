package entity

import "time"

type User struct {
	ID          int64     `json:"id"`
	TelegramID  int64     `json:"telegram_id"`
	Username    string    `json:"username"`
	Age         int       `json:"age"`
	Gender      string    `json:"gender"`
	Location    string    `json:"location"`
	Description string    `json:"description,omitempty"`
	PhotoURL    string    `json:"photo_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	IsVisible   bool      `json:"is_visible"`
}
