package config

type Settings struct {
	MaxPosts         int
	FetchLimit       int
	HoursWindow      int
	MinPoints        int
	MinComments      int
	CommentWeight    float64
	RecencyBonusMax  int
	MaxRootComments  int
	MaxChildComments int
	MaxCommentLevel  int
}

var DefaultSettings = Settings{
	MaxPosts:         24,
	FetchLimit:       200,
	HoursWindow:      24,
	MinPoints:        50,
	MinComments:      20,
	CommentWeight:    0.75,
	RecencyBonusMax:  100,
	MaxRootComments:  12,
	MaxChildComments: 8,
	MaxCommentLevel:  3,
}
