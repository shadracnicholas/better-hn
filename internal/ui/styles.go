package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorOrange = lipgloss.Color("#ff6600")
	colorBorder = lipgloss.Color("#333333")
	colorText   = lipgloss.Color("#d4d4d4")
	colorMuted  = lipgloss.Color("#666666")
	colorWhite  = lipgloss.Color("#ffffff")
	colorGray   = lipgloss.Color("#999999")
)

var (
	styleTitle        = lipgloss.NewStyle().Foreground(colorOrange).Bold(true)
	styleLoadingTitle = lipgloss.NewStyle().Foreground(colorOrange).Bold(true).Padding(0, 1)
	styleNormal       = lipgloss.NewStyle().Foreground(colorText)
	styleDim          = lipgloss.NewStyle().Foreground(colorMuted)
	styleAccent       = lipgloss.NewStyle().Foreground(colorOrange)
)

var (
	styleItemSelected     = lipgloss.NewStyle().Background(lipgloss.Color("#2a2a2a")).Foreground(colorWhite)
	styleItemSelectedMeta = lipgloss.NewStyle().Background(lipgloss.Color("#2a2a2a")).Foreground(colorGray)
	styleItemNormal       = lipgloss.NewStyle().Foreground(colorText)
	styleItemMeta         = lipgloss.NewStyle().Foreground(colorMuted)
)

var (
	styleBorderRight = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderRight(true).
				BorderForeground(colorBorder)

	styleBorderTop = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(colorBorder)
)
