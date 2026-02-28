package renderer

import (
	"fmt"
	"strings"
	"time"

	"github-stats-dashboard/api"
)

// ANSI color codes
const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Dim     = "\033[2m"

	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	BgBlack = "\033[40m"
	BgBlue  = "\033[44m"
)

const width = 60

func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

func HideCursor() {
	fmt.Print("\033[?25l")
}

func ShowCursor() {
	fmt.Print("\033[?25h")
}

func moveTo(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

func box(title, color string, lines []string) string {
	inner := width - 2
	var sb strings.Builder

	// Top border
	sb.WriteString(color + "┌─ " + Bold + title + Reset + color)
	titleLen := len(title) + 3
	sb.WriteString(strings.Repeat("─", inner-titleLen))
	sb.WriteString("┐" + Reset + "\n")

	// Lines
	for _, l := range lines {
		visible := visibleLen(l)
		padding := inner - visible - 1
		if padding < 0 {
			padding = 0
		}
		sb.WriteString(color + "│ " + Reset + l + strings.Repeat(" ", padding) + color + "│" + Reset + "\n")
	}

	// Bottom border
	sb.WriteString(color + "└" + strings.Repeat("─", inner) + "┘" + Reset + "\n")

	return sb.String()
}

// visibleLen strips ANSI codes and returns visible character count
func visibleLen(s string) int {
	inEsc := false
	count := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if c == 'm' {
				inEsc = false
			}
			continue
		}
		count++
	}
	return count
}

func colorize(s, color string) string {
	return color + s + Reset
}

func commitBar(count, max int) string {
	barWidth := 20
	if max == 0 {
		return colorize(strings.Repeat("░", barWidth), Dim)
	}
	filled := (count * barWidth) / max
	if filled == 0 && count > 0 {
		filled = 1
	}
	bar := colorize(strings.Repeat("█", filled), Green)
	bar += colorize(strings.Repeat("░", barWidth-filled), Dim)
	return bar
}

func streakBadge(streak int) string {
	if streak == 0 {
		return colorize("● no streak", Dim)
	}
	if streak >= 7 {
		return colorize(fmt.Sprintf("🔥 %dd streak — ON FIRE!", streak), Yellow+Bold)
	}
	return colorize(fmt.Sprintf("⚡ %d day streak", streak), Yellow)
}

func Render(s *api.Stats) {
	ClearScreen()
	moveTo(1, 1)

	out := ""

	// ── Header ──────────────────────────────────────────────────
	now := s.FetchedAt
	dateStr := now.Format("Monday, Jan 02 2006  15:04:05")
	headerLines := []string{
		fmt.Sprintf("%s  %s", colorize("@"+s.Username, Cyan+Bold), colorize(s.Name, White)),
		colorize(dateStr, Dim),
		fmt.Sprintf("%s  %s  %s",
			colorize(fmt.Sprintf("repos: %d", s.TotalRepos), Blue),
			colorize(fmt.Sprintf("followers: %d", s.Followers), Blue),
			colorize(fmt.Sprintf("runtime: %s/%s", s.GoVersion, s.CPUArch), Dim),
		),
	}
	out += box("GitHub Stats Dashboard", Blue, headerLines)

	// ── Today's Activity ────────────────────────────────────────
	todayLabel := colorize(fmt.Sprintf("%d commits today", s.TodayCommits), Green+Bold)
	monthLabel := colorize(fmt.Sprintf("%d this month", s.MonthCommits), Cyan)
	activityLines := []string{
		fmt.Sprintf("  %s   %s", todayLabel, monthLabel),
		"",
		"  " + streakBadge(s.Streak),
	}
	out += box("Activity", Green, activityLines)

	// ── 7-Day Commit Graph ───────────────────────────────────────
	maxCommits := 0
	for _, v := range s.CommitGraph {
		if v > maxCommits {
			maxCommits = v
		}
	}

	days := []string{"Today", "  -1d", "  -2d", "  -3d", "  -4d", "  -5d", "  -6d"}
	graphLines := []string{}
	for i := 0; i < 7; i++ {
		cnt := s.CommitGraph[i]
		bar := commitBar(cnt, maxCommits)
		label := colorize(days[i], White)
		cntStr := colorize(fmt.Sprintf("%3d", cnt), Green)
		graphLines = append(graphLines, fmt.Sprintf(" %s │ %s %s", label, bar, cntStr))
	}
	out += box("Commit Graph (last 7 days)", Cyan, graphLines)

	// ── Top Repos ────────────────────────────────────────────────
	repoLines := []string{}
	limit := 8
	if len(s.RepoCommits) < limit {
		limit = len(s.RepoCommits)
	}
	if limit == 0 {
		repoLines = append(repoLines, colorize("  No commits in active repos this month.", Dim))
	}
	for i := 0; i < limit; i++ {
		rc := s.RepoCommits[i]
		todayMark := ""
		if rc.Today > 0 {
			todayMark = colorize(fmt.Sprintf(" [+%d today]", rc.Today), Green)
		}
		name := colorize(rc.Name, White+Bold)
		cnt := colorize(fmt.Sprintf("%d commits", rc.Commits), Cyan)

		// Truncate long repo names
		displayName := rc.Name
		if len(displayName) > 22 {
			displayName = displayName[:19] + "..."
		}
		name = colorize(displayName, White+Bold)

		repoLines = append(repoLines, fmt.Sprintf("  %-22s %s%s", name, cnt, todayMark))
	}
	out += box("Top Repos (this month)", Magenta, repoLines)

	// ── Footer ───────────────────────────────────────────────────
	nextRefresh := s.FetchedAt.Add(60 * time.Second)
	remaining := time.Until(nextRefresh).Round(time.Second)
	footerLines := []string{
		fmt.Sprintf("  %s   %s",
			colorize("Next refresh in "+remaining.String(), Dim),
			colorize("Ctrl+C to exit", Dim),
		),
	}
	out += box("", Dim, footerLines)

	fmt.Print(out)
}

func RenderError(err error) {
	ClearScreen()
	moveTo(1, 1)
	lines := []string{
		colorize("  Failed to fetch GitHub stats:", Red+Bold),
		colorize("  "+err.Error(), Yellow),
		"",
		colorize("  Check your GITHUB_TOKEN and internet connection.", Dim),
		colorize("  Retrying in 60 seconds...", Dim),
	}
	fmt.Print(box("Error", Red, lines))
}
