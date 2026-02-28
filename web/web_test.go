package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github-stats-dashboard/api"
)

func TestPrepareDashData(t *testing.T) {
	stats := &api.Stats{
		Username:     "testuser",
		Name:         "Test User",
		FetchedAt:    time.Date(2026, 2, 28, 12, 0, 0, 0, time.UTC),
		TotalRepos:   10,
		PublicRepos:  8,
		Followers:    42,
		TodayCommits: 5,
		MonthCommits: 30,
		Streak:       3,
		CommitGraph:  [7]int{5, 3, 2, 0, 8, 1, 4},
		RepoCommits: []api.RepoCommit{
			{Name: "repo-a", Commits: 15, Today: 3},
			{Name: "repo-b", Commits: 10, Today: 0},
		},
		GoVersion: "go1.21",
		CPUArch:   "amd64",
	}

	data := prepareDashData(stats)

	if data.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %q", data.Username)
	}
	if data.TodayCommits != 5 {
		t.Errorf("expected 5 today commits, got %d", data.TodayCommits)
	}
	if data.MonthCommits != 30 {
		t.Errorf("expected 30 month commits, got %d", data.MonthCommits)
	}
	if data.Streak != 3 {
		t.Errorf("expected streak 3, got %d", data.Streak)
	}
	if !strings.Contains(data.StreakLabel, "3 day streak") {
		t.Errorf("expected streak label to mention '3 day streak', got %q", data.StreakLabel)
	}
	if len(data.Graph) != 7 {
		t.Fatalf("expected 7 graph entries, got %d", len(data.Graph))
	}
	// Max is 8 (index 4), so index 0 (count=5) => percent = 5*100/8 = 62
	if data.Graph[0].Percent != 62 {
		t.Errorf("expected graph[0] percent 62, got %d", data.Graph[0].Percent)
	}
	if len(data.TopRepos) != 2 {
		t.Fatalf("expected 2 top repos, got %d", len(data.TopRepos))
	}
	if data.TopRepos[0].Name != "repo-a" {
		t.Errorf("expected first repo 'repo-a', got %q", data.TopRepos[0].Name)
	}
}

func TestHandleDashboard_WithStats(t *testing.T) {
	s := &Server{
		stats: &api.Stats{
			Username:     "testuser",
			Name:         "Test User",
			FetchedAt:    time.Date(2026, 2, 28, 12, 0, 0, 0, time.UTC),
			TotalRepos:   5,
			Followers:    10,
			TodayCommits: 2,
			MonthCommits: 20,
			CommitGraph:  [7]int{2, 1, 0, 0, 0, 0, 0},
			GoVersion:    "go1.21",
			CPUArch:      "amd64",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	s.handleDashboard(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "@testuser") {
		t.Error("response should contain @testuser")
	}
	if !strings.Contains(body, "GitHub Stats Dashboard") {
		t.Error("response should contain dashboard title")
	}
	if !strings.Contains(body, "commits today") {
		t.Error("response should contain today commits label")
	}
}

func TestHandleDashboard_Loading(t *testing.T) {
	s := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	s.handleDashboard(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "Loading") {
		t.Error("expected loading page when stats is nil")
	}
}

func TestHandleDashboard_Error(t *testing.T) {
	s := &Server{
		err: http.ErrServerClosed,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	s.handleDashboard(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "Error") {
		t.Error("expected error page when err is set")
	}
}

func TestStreakLabels(t *testing.T) {
	tests := []struct {
		streak int
		want   string
	}{
		{0, "no streak"},
		{3, "3 day streak"},
		{7, "ON FIRE!"},
	}
	for _, tt := range tests {
		stats := &api.Stats{Streak: tt.streak, FetchedAt: time.Now(), CommitGraph: [7]int{}, GoVersion: "go1.21", CPUArch: "amd64"}
		data := prepareDashData(stats)
		if !strings.Contains(data.StreakLabel, tt.want) {
			t.Errorf("streak %d: expected label containing %q, got %q", tt.streak, tt.want, data.StreakLabel)
		}
	}
}
