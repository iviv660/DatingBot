package entity

import "time"

type Match struct {
	ID        int64     `json:"id"`
	FromUser  int64     `json:"from_user"` // кто поставил лайк
	ToUser    int64     `json:"to_user"`   // кому
	IsLike    bool      `json:"is_like"`   // true=лайк, false=дизлайк
	CreatedAt time.Time `json:"created_at"`
}
