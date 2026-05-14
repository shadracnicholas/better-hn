package api

type Post struct {
	ID            int       `json:"id"`
	Title         string    `json:"title"`
	Points        *int      `json:"points"`
	User          *string   `json:"user"`
	Time          int64     `json:"time"`
	TimeAgo       string    `json:"time_ago"`
	Type          string    `json:"type"`
	Content       string    `json:"content"`
	URL           string    `json:"url"`
	Domain        string    `json:"domain"`
	CommentsCount int       `json:"comments_count"`
	Comments      []Comment `json:"comments"`
}

type Comment struct {
	ID            int       `json:"id"`
	User          *string   `json:"user"`
	Content       *string   `json:"content"`
	Level         int       `json:"level"`
	Time          int64     `json:"time"`
	TimeAgo       string    `json:"time_ago"`
	CommentsCount int       `json:"comments_count"`
	Comments      []Comment `json:"comments"`
	Dead          bool      `json:"dead"`
	Deleted       bool      `json:"deleted"`
}
