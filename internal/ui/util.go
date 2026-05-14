package ui

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	rePreBlock = regexp.MustCompile(`(?is)<pre[^>]*>(.*?)</pre>`)
	reAnchor   = regexp.MustCompile(`(?i)<a\s+[^>]*href=["']([^"']+)["'][^>]*>(.*?)</a>`)
	reItalic   = regexp.MustCompile(`(?is)<(?:i|em)>(.*?)</(?:i|em)>`)
	reCode     = regexp.MustCompile(`(?is)<code>(.*?)</code>`)
	reTag      = regexp.MustCompile(`<[^>]+>`)
)

func htmlToText(html string) string {
	s := html
	s = strings.ReplaceAll(s, "<p>", "\n\n")
	s = strings.ReplaceAll(s, "</p>", "")

	s = rePreBlock.ReplaceAllStringFunc(s, func(m string) string {
		inner := rePreBlock.FindStringSubmatch(m)
		if inner == nil {
			return m
		}
		return "\n" + inner[1] + "\n"
	})

	s = reAnchor.ReplaceAllStringFunc(s, func(m string) string {
		parts := reAnchor.FindStringSubmatch(m)
		if parts == nil {
			return m
		}
		url := parts[1]
		text := strings.TrimSpace(reTag.ReplaceAllString(parts[2], ""))
		if text == url {
			return text
		}
		return text + " (" + url + ")"
	})

	s = reItalic.ReplaceAllString(s, "$1")

	s = reCode.ReplaceAllStringFunc(s, func(m string) string {
		parts := reCode.FindStringSubmatch(m)
		if parts == nil {
			return m
		}
		return "`" + parts[1] + "`"
	})

	s = reTag.ReplaceAllString(s, "")
	s = strings.NewReplacer(
		"&amp;", "&", "&lt;", "<", "&gt;", ">",
		"&quot;", `"`, "&#x27;", "'", "&#x2F;", "/",
		"&#x2f;", "/", "&#39;", "'", "&nbsp;", " ",
		"&mdash;", "—", "&ndash;", "–", "&hellip;", "…",
	).Replace(s)

	return strings.TrimSpace(s)
}

func wordWrap(text string, width int) []string {
	if text == "" {
		return nil
	}
	if width <= 0 {
		return []string{text}
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	current := ""
	currentLen := 0
	for _, word := range words {
		wordLen := utf8.RuneCountInString(word)
		if currentLen == 0 {
			current = word
			currentLen = wordLen
			continue
		}
		if currentLen+1+wordLen <= width {
			current += " " + word
			currentLen += 1 + wordLen
		} else {
			lines = append(lines, current)
			current = word
			currentLen = wordLen
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 1 {
		return string(runes[:n])
	}
	return string(runes[:n-1]) + "…"
}
