package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sort"
	"time"
)

const baseURL = "https://api.github.com"

type Client struct {
	token    string
	username string
	http     *http.Client
}

type Stats struct {
	Username          string
	Name              string
	FetchedAt         time.Time
	TotalRepos        int
	PublicRepos       int
	Followers         int
	TodayCommits      int
	MonthCommits      int
	Streak            int
	RepoCommits       []RepoCommit
	CommitGraph       [7]int // last 7 days
	CPUArch           string
	GoVersion         string
}

type RepoCommit struct {
	Name    string
	Commits int
	Today   int
}

func NewClient(token, username string) *Client {
	return &Client{
		token:    token,
		username: username,
		http:     &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) get(url string, v interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
	}
	return json.Unmarshal(body, v)
}

func (c *Client) FetchStats() (*Stats, error) {
	stats := &Stats{
		FetchedAt: time.Now(),
		CPUArch:   runtime.GOARCH,
		GoVersion: runtime.Version(),
	}

	// 1. Fetch authenticated user
	var user struct {
		Login     string `json:"login"`
		Name      string `json:"name"`
		Repos     int    `json:"public_repos"`
		Followers int    `json:"followers"`
	}
	if err := c.get(baseURL+"/user", &user); err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}
	stats.Username = user.Login
	stats.Name = user.Name
	stats.PublicRepos = user.Repos
	stats.Followers = user.Followers

	if c.username == "" {
		c.username = user.Login
	}

	// 2. Fetch owned repos (up to 100)
	var repos []struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		Fork bool `json:"fork"`
	}
	repoURL := fmt.Sprintf("%s/user/repos?per_page=100&type=owner&sort=pushed", baseURL)
	if err := c.get(repoURL, &repos); err != nil {
		return nil, fmt.Errorf("fetching repos: %w", err)
	}

	stats.TotalRepos = len(repos)

	// 3. For each repo, fetch commits by authenticated user in last 30 days
	now := time.Now()
	monthStart := now.AddDate(0, -1, 0)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Build 7-day windows
	dayStarts := [8]time.Time{}
	for i := 0; i <= 7; i++ {
		d := now.AddDate(0, 0, -i)
		dayStarts[i] = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
	}

	type repoResult struct {
		name        string
		todayCount  int
		monthCount  int
		dailyCounts [7]int
	}

	results := make([]repoResult, 0, len(repos))

	// Limit to 20 most recently pushed repos to avoid rate limits
	limit := len(repos)
	if limit > 20 {
		limit = 20
	}

	for _, repo := range repos[:limit] {
		if repo.Fork {
			continue
		}

		var commits []struct {
			Commit struct {
				Author struct {
					Date string `json:"date"`
				} `json:"author"`
			} `json:"commit"`
		}

		url := fmt.Sprintf("%s/repos/%s/%s/commits?author=%s&since=%s&per_page=100",
			baseURL, c.username, repo.Name, c.username,
			monthStart.Format(time.RFC3339))

		if err := c.get(url, &commits); err != nil {
			// 409 = empty repo, 422 = git error — skip silently
			continue
		}

		rr := repoResult{name: repo.Name}
		for _, cm := range commits {
			t, err := time.Parse(time.RFC3339, cm.Commit.Author.Date)
			if err != nil {
				continue
			}
			rr.monthCount++
			if t.After(todayStart) {
				rr.todayCount++
			}
			// Which day bucket?
			for i := 0; i < 7; i++ {
				if t.After(dayStarts[i+1]) && (t.Before(dayStarts[i]) || t.Equal(dayStarts[i])) {
					rr.dailyCounts[i]++
					break
				}
			}
		}

		if rr.monthCount > 0 {
			results = append(results, rr)
		}
	}

	// Aggregate
	for _, rr := range results {
		stats.TodayCommits += rr.todayCount
		stats.MonthCommits += rr.monthCount
		for i := 0; i < 7; i++ {
			stats.CommitGraph[i] += rr.dailyCounts[i]
		}
		stats.RepoCommits = append(stats.RepoCommits, RepoCommit{
			Name:    rr.name,
			Commits: rr.monthCount,
			Today:   rr.todayCount,
		})
	}

	// Sort repos by month commits desc
	sort.Slice(stats.RepoCommits, func(i, j int) bool {
		return stats.RepoCommits[i].Commits > stats.RepoCommits[j].Commits
	})

	// Calculate streak (consecutive days with commits, going back)
	stats.Streak = 0
	for i := 0; i < 7; i++ {
		if stats.CommitGraph[i] > 0 {
			stats.Streak++
		} else {
			break
		}
	}

	return stats, nil
}
