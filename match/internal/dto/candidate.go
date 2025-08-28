package dto

type Candidate struct {
	TargetGender string  `json:"target_gender"`
	MinAge       int     `json:"min_age"`
	MaxAge       int     `json:"max_age"`
	Location     string  `json:"location"`
	Limit        int     `json:"limit"`
	ExcludeIDs   []int64 `json:"exclude_ids"`
}
