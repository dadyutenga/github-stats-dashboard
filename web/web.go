package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github-stats-dashboard/api"
)

// Server holds the web server state.
type Server struct {
	client *api.Client
	mu     sync.RWMutex
	stats  *api.Stats
	err    error
}

// NewServer creates a new web dashboard server.
func NewServer(client *api.Client) *Server {
	return &Server{client: client}
}

// Start launches the web server on the given address and begins
// a background refresh loop that polls the GitHub API every 60 seconds.
func (s *Server) Start(addr string) error {
	// Initial fetch
	s.refresh()

	// Background refresh loop
	go func() {
		for {
			time.Sleep(60 * time.Second)
			s.refresh()
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleDashboard)

	log.Printf("🌐 Dashboard running at http://%s\n", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) refresh() {
	stats, err := s.client.FetchStats()
	s.mu.Lock()
	s.stats = stats
	s.err = err
	s.mu.Unlock()
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	stats := s.stats
	fetchErr := s.err
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if fetchErr != nil {
		fmt.Fprintf(w, errorHTML, fetchErr.Error())
		return
	}

	if stats == nil {
		fmt.Fprint(w, loadingHTML)
		return
	}

	data := prepareDashData(stats)
	if err := dashTmpl.Execute(w, data); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// dashData carries template-ready values.
type dashData struct {
	Username     string
	Name         string
	FetchedAt    string
	TotalRepos   int
	Followers    int
	GoVersion    string
	CPUArch      string
	TodayCommits int
	MonthCommits int
	Streak       int
	StreakLabel  string
	Graph        []graphDay
	TopRepos     []repoRow
}

type graphDay struct {
	Label   string
	Count   int
	Percent int // 0-100
}

type repoRow struct {
	Name    string
	Commits int
	Today   int
}

func prepareDashData(s *api.Stats) dashData {
	d := dashData{
		Username:     s.Username,
		Name:         s.Name,
		FetchedAt:    s.FetchedAt.Format("Monday, Jan 02 2006  15:04:05"),
		TotalRepos:   s.TotalRepos,
		Followers:    s.Followers,
		GoVersion:    s.GoVersion,
		CPUArch:      s.CPUArch,
		TodayCommits: s.TodayCommits,
		MonthCommits: s.MonthCommits,
		Streak:       s.Streak,
	}

	switch {
	case s.Streak == 0:
		d.StreakLabel = "● no streak"
	case s.Streak >= 7:
		d.StreakLabel = fmt.Sprintf("🔥 %d day streak — ON FIRE!", s.Streak)
	default:
		d.StreakLabel = fmt.Sprintf("⚡ %d day streak", s.Streak)
	}

	maxCommits := 0
	for _, v := range s.CommitGraph {
		if v > maxCommits {
			maxCommits = v
		}
	}
	labels := []string{"Today", "-1d", "-2d", "-3d", "-4d", "-5d", "-6d"}
	for i := 0; i < 7; i++ {
		pct := 0
		if maxCommits > 0 {
			pct = (s.CommitGraph[i] * 100) / maxCommits
		}
		d.Graph = append(d.Graph, graphDay{Label: labels[i], Count: s.CommitGraph[i], Percent: pct})
	}

	limit := 8
	if len(s.RepoCommits) < limit {
		limit = len(s.RepoCommits)
	}
	for i := 0; i < limit; i++ {
		rc := s.RepoCommits[i]
		d.TopRepos = append(d.TopRepos, repoRow{Name: rc.Name, Commits: rc.Commits, Today: rc.Today})
	}

	return d
}

const loadingHTML = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>GitHub Stats Dashboard</title>
<meta http-equiv="refresh" content="3">
<style>body{background:#0d1117;color:#c9d1d9;font-family:monospace;display:flex;align-items:center;justify-content:center;height:100vh;margin:0}
</style></head><body><h2>⏳ Loading stats…</h2></body></html>`

const errorHTML = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>GitHub Stats Dashboard — Error</title>
<meta http-equiv="refresh" content="10">
<style>body{background:#0d1117;color:#f85149;font-family:monospace;padding:2em}</style>
</head><body><h2>❌ Error fetching stats</h2><pre>%s</pre><p style="color:#8b949e">Retrying…</p></body></html>`

var dashTmpl = template.Must(template.New("dash").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>GitHub Stats Dashboard — @{{.Username}}</title>
<meta http-equiv="refresh" content="60">
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{background:#0d1117;color:#c9d1d9;font-family:'Segoe UI',system-ui,-apple-system,sans-serif;padding:1.5rem}
.container{max-width:720px;margin:0 auto}
.card{border:1px solid #30363d;border-radius:8px;padding:1rem 1.25rem;margin-bottom:1rem;background:#161b22}
.card h2{font-size:.85rem;text-transform:uppercase;letter-spacing:.08em;color:#8b949e;margin-bottom:.75rem;border-bottom:1px solid #21262d;padding-bottom:.5rem}
.header-info{display:flex;align-items:baseline;gap:.75rem;flex-wrap:wrap}
.username{color:#58a6ff;font-size:1.25rem;font-weight:700}
.fullname{color:#c9d1d9;font-size:1rem}
.meta{color:#8b949e;font-size:.85rem;margin-top:.35rem}
.meta span{margin-right:1.25rem}
.meta .blue{color:#58a6ff}
.stat-row{display:flex;gap:2rem;flex-wrap:wrap;align-items:baseline}
.stat-big{font-size:1.8rem;font-weight:700;color:#3fb950}
.stat-big .label{font-size:.85rem;font-weight:400;color:#8b949e}
.stat-secondary{font-size:1.1rem;color:#58a6ff}
.stat-secondary .label{font-size:.85rem;color:#8b949e}
.streak{margin-top:.5rem;color:#d29922;font-size:.95rem}
.graph-row{display:flex;align-items:center;margin-bottom:.35rem;font-size:.85rem}
.graph-label{width:3rem;text-align:right;margin-right:.75rem;color:#8b949e}
.graph-bar-bg{flex:1;height:18px;background:#21262d;border-radius:3px;overflow:hidden}
.graph-bar{height:100%;background:#3fb950;border-radius:3px;transition:width .3s}
.graph-count{width:2.5rem;text-align:right;margin-left:.5rem;color:#3fb950;font-weight:600}
.repo-item{display:flex;justify-content:space-between;align-items:center;padding:.4rem 0;border-bottom:1px solid #21262d;font-size:.9rem}
.repo-item:last-child{border-bottom:none}
.repo-name{color:#c9d1d9;font-weight:600}
.repo-stats{color:#58a6ff}
.repo-today{color:#3fb950;margin-left:.5rem;font-size:.8rem}
.footer{text-align:center;color:#484f58;font-size:.8rem;margin-top:.5rem}
.empty{color:#484f58;font-style:italic}
</style>
</head>
<body>
<div class="container">

<div class="card">
  <h2>GitHub Stats Dashboard</h2>
  <div class="header-info">
    <span class="username">@{{.Username}}</span>
    <span class="fullname">{{.Name}}</span>
  </div>
  <div class="meta">
    <span>{{.FetchedAt}}</span><br>
    <span class="blue">repos: {{.TotalRepos}}</span>
    <span class="blue">followers: {{.Followers}}</span>
    <span>runtime: {{.GoVersion}}/{{.CPUArch}}</span>
  </div>
</div>

<div class="card">
  <h2>Activity</h2>
  <div class="stat-row">
    <div><span class="stat-big">{{.TodayCommits}}</span> <span class="stat-big label">commits today</span></div>
    <div><span class="stat-secondary">{{.MonthCommits}}</span> <span class="stat-secondary label">this month</span></div>
  </div>
  <div class="streak">{{.StreakLabel}}</div>
</div>

<div class="card">
  <h2>Commit Graph (last 7 days)</h2>
  {{range .Graph}}
  <div class="graph-row">
    <span class="graph-label">{{.Label}}</span>
    <div class="graph-bar-bg"><div class="graph-bar" style="width:{{.Percent}}%"></div></div>
    <span class="graph-count">{{.Count}}</span>
  </div>
  {{end}}
</div>

<div class="card">
  <h2>Top Repos (this month)</h2>
  {{if .TopRepos}}
    {{range .TopRepos}}
    <div class="repo-item">
      <span class="repo-name">{{.Name}}</span>
      <span><span class="repo-stats">{{.Commits}} commits</span>{{if gt .Today 0}}<span class="repo-today">+{{.Today}} today</span>{{end}}</span>
    </div>
    {{end}}
  {{else}}
    <div class="empty">No commits in active repos this month.</div>
  {{end}}
</div>

<div class="footer">Auto-refreshes every 60 seconds · Powered by Go</div>

</div>
</body>
</html>
`))
