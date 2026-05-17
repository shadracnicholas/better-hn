package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/shadracnicholas/better-hn/internal/config"
)

const (
	topStoriesURL = "https://hacker-news.firebaseio.com/v0/topstories.json"
	itemURLFmt    = "https://api.hnpwa.com/v0/item/%d.json"
)

var client = &http.Client{Timeout: 15 * time.Second}

func GetTopStoryIDs() ([]int, error) {
	resp, err := client.Get(topStoriesURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var ids []int
	if err := json.NewDecoder(resp.Body).Decode(&ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func GetPostByID(id int, s config.Settings) (*Post, error) {
	resp, err := client.Get(fmt.Sprintf(itemURLFmt, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var post Post
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return nil, err
	}
	post.Comments = trimComments(post.Comments, s, 0)
	return &post, nil
}

func GetRankedPosts(s config.Settings) ([]Post, error) {
	ids, err := GetTopStoryIDs()
	if err != nil {
		return nil, err
	}

	if len(ids) > s.FetchLimit {
		ids = ids[:s.FetchLimit]
	}

	sem := make(chan struct{}, 20)
	var (
		mu    sync.Mutex
		posts []Post
		wg    sync.WaitGroup
	)

	for _, id := range ids {
		id := id
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			resp, err := client.Get(fmt.Sprintf(itemURLFmt, id))
			if err != nil {
				return
			}
			defer resp.Body.Close()

			var post Post
			if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
				return
			}
			post.Comments = nil

			mu.Lock()
			posts = append(posts, post)
			mu.Unlock()
		}()
	}
	wg.Wait()

	now := time.Now().Unix()
	var filtered []Post
	for _, post := range posts {
		if post.Type != "link" {
			continue
		}
		if post.Time < now-int64(s.HoursWindow)*3600 {
			continue
		}
		pts := 0
		if post.Points != nil {
			pts = *post.Points
		}
		if pts < s.MinPoints && post.CommentsCount < s.MinComments {
			continue
		}
		filtered = append(filtered, post)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return postScore(filtered[i], s, now) > postScore(filtered[j], s, now)
	})

	if len(filtered) > s.MaxPosts {
		filtered = filtered[:s.MaxPosts]
	}
	return filtered, nil
}

func postScore(post Post, s config.Settings, now int64) float64 {
	pts := 0
	if post.Points != nil {
		pts = *post.Points
	}
	hoursOld := float64(now-post.Time) / 3600.0
	recency := float64(s.RecencyBonusMax) * (1 - hoursOld/float64(s.HoursWindow))
	if recency < 0 {
		recency = 0
	}
	return float64(pts) + float64(post.CommentsCount)*s.CommentWeight + recency
}

func trimComments(comments []Comment, s config.Settings, depth int) []Comment {
	limit := s.MaxRootComments
	if depth > 0 {
		limit = s.MaxChildComments
	}
	if len(comments) > limit {
		comments = comments[:limit]
	}
	if depth >= s.MaxCommentLevel {
		for i := range comments {
			comments[i].Comments = nil
		}
		return comments
	}
	for i := range comments {
		comments[i].Comments = trimComments(comments[i].Comments, s, depth+1)
	}
	return comments
}
