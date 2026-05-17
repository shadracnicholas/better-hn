package ui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shadracnicholas/better-hn/internal/api"
	"github.com/shadracnicholas/better-hn/internal/config"
)

type postsLoadedMsg struct {
	posts []api.Post
	err   error
}

type postDetailMsg struct {
	postID int
	post   *api.Post
	err    error
}

type Model struct {
	width, height int

	posts       []api.Post
	selectedIdx int
	loading     bool
	loadErr     error
	lastFetch   time.Time

	fullPost         *api.Post
	fullPostLoading  bool
	viewport         viewport.Model
	rootCommentLines []int
	rootCommentIdx   int

	spinner      spinner.Model
	blinkSpinner spinner.Model
}

var cursorBlink = spinner.Spinner{
	Frames: []string{"█", " "},
	FPS:    time.Second / 2,
}

func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styleAccent

	b := spinner.New()
	b.Spinner = cursorBlink
	b.Style = styleAccent

	return Model{
		spinner:      s,
		blinkSpinner: b,
		loading:      true,
		viewport:     viewport.New(0, 0),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.blinkSpinner.Tick, fetchPostsCmd(config.DefaultSettings))
}

func fetchPostsCmd(s config.Settings) tea.Cmd {
	return func() tea.Msg {
		posts, err := api.GetRankedPosts(s)
		return postsLoadedMsg{posts: posts, err: err}
	}
}

func fetchDetailCmd(id int, s config.Settings) tea.Cmd {
	return func() tea.Msg {
		post, err := api.GetPostByID(id, s)
		return postDetailMsg{postID: id, post: post, err: err}
	}
}

func openURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		default:
			cmd = exec.Command("xdg-open", url)
		}
		_ = cmd.Start()
		return nil
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = m.detailWidth()
		m.viewport.Height = m.height - 2
		if m.fullPost != nil {
			m.refreshDetail()
		}

	case postsLoadedMsg:
		m.loading = false
		m.lastFetch = time.Now()
		if msg.err != nil {
			m.loadErr = msg.err
			return m, nil
		}
		m.posts = msg.posts
		m.selectedIdx = 0
		if len(m.posts) > 0 {
			m.fullPostLoading = true
			return m, fetchDetailCmd(m.posts[0].ID, config.DefaultSettings)
		}

	case postDetailMsg:
		if len(m.posts) > 0 && msg.postID != m.posts[m.selectedIdx].ID {
			return m, nil
		}
		m.fullPostLoading = false
		if msg.err != nil {
			return m, nil
		}
		m.fullPost = msg.post
		m.rootCommentIdx = 0
		m.refreshDetail()
		m.viewport.GotoTop()

	case spinner.TickMsg:
		var cmd1, cmd2 tea.Cmd
		m.spinner, cmd1 = m.spinner.Update(msg)
		m.blinkSpinner, cmd2 = m.blinkSpinner.Update(msg)
		return m, tea.Batch(cmd1, cmd2)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "j":
			if !m.loading && len(m.posts) > 0 && m.selectedIdx < len(m.posts)-1 {
				m.selectedIdx++
				m.fullPostLoading = true
				m.fullPost = nil
				return m, fetchDetailCmd(m.posts[m.selectedIdx].ID, config.DefaultSettings)
			}

		case "k":
			if !m.loading && m.selectedIdx > 0 {
				m.selectedIdx--
				m.fullPostLoading = true
				m.fullPost = nil
				return m, fetchDetailCmd(m.posts[m.selectedIdx].ID, config.DefaultSettings)
			}

		case " ":
			if len(m.rootCommentLines) > 0 {
				m.rootCommentIdx = (m.rootCommentIdx + 1) % len(m.rootCommentLines)
				m.viewport.SetYOffset(m.rootCommentLines[m.rootCommentIdx])
			}

		case "o":
			url := ""
			if m.fullPost != nil {
				url = m.fullPost.URL
			} else if len(m.posts) > 0 {
				url = m.posts[m.selectedIdx].URL
			}
			if url != "" {
				return m, openURLCmd(url)
			}

		case "r":
			m.posts = nil
			m.fullPost = nil
			m.loadErr = nil
			m.loading = true
			return m, tea.Batch(m.spinner.Tick, m.blinkSpinner.Tick, fetchPostsCmd(config.DefaultSettings))

		case "up", "down", "pgup", "pgdown", "ctrl+u", "ctrl+d":
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *Model) refreshDetail() {
	content, lines := RenderDetail(m.fullPost, m.detailWidth()-4)
	m.viewport.SetContent(content)
	m.rootCommentLines = lines
}

func (m Model) listWidth() int {
	w := m.width * 35 / 100
	if w < 25 {
		w = 25
	}
	if w > 60 {
		w = 60
	}
	return w
}

func (m Model) detailWidth() int {
	return m.width - m.listWidth()
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		m.renderHeader(),
		m.renderContent(),
		m.renderShortcuts(),
	)
}

func (m Model) renderHeader() string {
	title := styleTitle.Render(" HN ")
	var status string
	switch {
	case m.loading:
		status = m.spinner.View() + " fetching…"
	case m.loadErr != nil:
		status = m.loadErr.Error()
	default:
		ago := time.Since(m.lastFetch).Round(time.Minute)
		agoStr := strings.TrimSuffix(ago.String(), "0s")
		if agoStr == "" {
			agoStr = "<1m"
		}
		status = fmt.Sprintf("%d stories · %s ago", len(m.posts), agoStr)
	}
	return title + styleDim.Render(status)
}

func (m Model) renderContent() string {
	if m.loading {
		text := styleLoadingTitle.Render("better hn") + styleAccent.Render(m.blinkSpinner.View())
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height-2).
			Align(lipgloss.Center, lipgloss.Center).
			Render(text)
	}

	storyList := RenderStoryList(m.posts, m.selectedIdx, m.listWidth(), m.height-2)

	var detailPanel string
	if m.fullPostLoading {
		loading := styleDim.Render(m.spinner.View() + " Loading…")
		detailPanel = lipgloss.NewStyle().
			Width(m.detailWidth()).
			Height(m.height - 2).
			Render(loading)
	} else {
		detailPanel = lipgloss.NewStyle().
			Width(m.detailWidth()).
			Height(m.height-2).
			Padding(0, 2).
			Render(m.viewport.View())
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, storyList, detailPanel)
}

func (m Model) renderShortcuts() string {
	shortcuts := "j/k navigate · space next comment · o open · r refresh · q quit"
	return styleBorderTop.Width(m.width).Align(lipgloss.Center).Render(styleDim.Render(shortcuts))
}
