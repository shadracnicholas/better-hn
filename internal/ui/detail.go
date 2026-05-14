package ui

import (
	"fmt"
	"strings"

	"github.com/shadracnicholas/better-hn/internal/api"
)

func RenderDetail(post *api.Post, width int) (string, []int) {
	var lines []string
	var rootLines []int

	for _, l := range wordWrap(post.Title, width) {
		lines = append(lines, styleTitle.Render(l))
	}

	if post.Domain != "" {
		lines = append(lines, styleDim.Render(truncate(post.Domain, width)))
	} else {
		lines = append(lines, styleDim.Render(truncate(post.URL, width)))
	}

	lines = append(lines, "")

	points := 0
	if post.Points != nil {
		points = *post.Points
	}
	lines = append(lines, styleDim.Render(fmt.Sprintf("▲ %d points · %d comments · %s", points, post.CommentsCount, post.TimeAgo)))

	if post.Content != "" {
		lines = append(lines, "")
		for i, para := range strings.Split(htmlToText(post.Content), "\n\n") {
			if i > 0 {
				lines = append(lines, "")
			}
			for _, l := range wordWrap(para, width) {
				lines = append(lines, styleNormal.Render(l))
			}
		}
	}

	lines = append(lines, "")
	lines = append(lines, styleDim.Render(strings.Repeat("─", min(width, 60))))
	lines = append(lines, "")

	if len(post.Comments) == 0 {
		lines = append(lines, styleDim.Render("No comments"))
	} else {
		for i, c := range post.Comments {
			if i > 0 {
				lines = append(lines, "")
			}
			rootLines = append(rootLines, len(lines))
			lines = append(lines, renderComment(c, width)...)
		}
	}

	return strings.Join(lines, "\n"), rootLines
}

func renderComment(c api.Comment, width int) []string {
	indent := strings.Repeat("  ", c.Level)
	contentWidth := width - c.Level*2 - 2
	if contentWidth < 10 {
		contentWidth = 10
	}

	if c.Deleted || c.Dead {
		return []string{indent + styleDim.Render("[deleted]")}
	}

	var lines []string

	username := "[deleted]"
	if c.User != nil {
		username = *c.User
	}

	var userStyle string
	if c.Level == 0 {
		userStyle = styleAccent.Render(username)
	} else {
		userStyle = styleDim.Render(username)
	}
	lines = append(lines, indent+"│ "+userStyle+styleDim.Render(" · "+c.TimeAgo))

	if c.Content != nil && *c.Content != "" {
		for _, para := range strings.Split(htmlToText(*c.Content), "\n\n") {
			for _, l := range wordWrap(para, contentWidth) {
				lines = append(lines, indent+"│ "+styleNormal.Render(l))
			}
		}
	}

	for i, child := range c.Comments {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, renderComment(child, width)...)
	}

	return lines
}
