package dto

type UpdateProfileInput struct {
	Username    string `json:"username,omitempty"`
	Age         int    `json:"age,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Location    string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
	IsVisible   bool   `json:"is_visible,omitempty"`
}
