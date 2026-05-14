package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/shadracnicholas/better-hn/internal/api"
)

func RenderStoryList(posts []api.Post, selectedIdx, width, height int) string {
	innerWidth := width - 1
	if innerWidth < 1 {
		innerWidth = 1
	}

	visibleItems := height / 2
	if visibleItems < 1 {
		visibleItems = 1
	}

	offset := selectedIdx - visibleItems/2
	maxOffset := len(posts) - visibleItems
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}

	var lines []string

	end := offset + visibleItems
	if end > len(posts) {
		end = len(posts)
	}

	for i := offset; i < end; i++ {
		post := posts[i]
		selected := i == selectedIdx

		points := 0
		if post.Points != nil {
			points = *post.Points
		}

		indicator := "•"
		if selected {
			indicator = "›"
		}

		prefix1 := indicator + fmt.Sprintf("%2d", i+1) + ". "
		titleWidth := innerWidth - lipgloss.Width(prefix1)
		if titleWidth < 0 {
			titleWidth = 0
		}
		line1 := prefix1 + truncate(post.Title, titleWidth)

		meta := fmt.Sprintf("    ▲ %d · %d comments · %s", points, post.CommentsCount, post.TimeAgo)
		if post.Domain != "" {
			meta += " (" + post.Domain + ")"
		}
		line2 := truncate(meta, innerWidth)

		if selected {
			lines = append(lines, styleItemSelected.Width(innerWidth).Render(line1))
			lines = append(lines, styleItemSelectedMeta.Width(innerWidth).Render(line2))
		} else {
			lines = append(lines, styleItemNormal.Width(innerWidth).Render(line1))
			lines = append(lines, styleItemMeta.Width(innerWidth).Render(line2))
		}
	}

	return styleBorderRight.Width(innerWidth).Height(height).Render(strings.Join(lines, "\n"))
}
