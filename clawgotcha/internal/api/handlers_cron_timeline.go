package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/morpheumlabs/agentbook/clawgotcha/internal/db"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/httperr"
	"github.com/robfig/cron/v3"
)

// CronScheduleTimelineResponse is the GET /api/v1/cron-jobs/schedule-timeline payload.
type CronScheduleTimelineResponse struct {
	AsOf         time.Time                 `json:"as_of"`
	HorizonEnds  time.Time                 `json:"horizon_ends"`
	AnchoredBy   string                    `json:"anchored_by"`
	HorizonHours int                       `json:"horizon_hours"`
	MaxRuns      int                       `json:"max_runs"`
	Rows         []CronScheduleTimelineRow `json:"rows"`
}

// CronScheduleTimelineRow is one job plus projected 5-field cron instants in the window.
type CronScheduleTimelineRow struct {
	ID        string    `json:"ID"`
	Name      string    `json:"Name"`
	AgentName string    `json:"AgentName"`
	Schedule  string    `json:"Schedule"`
	Active    bool      `json:"Active"`
	UpdatedAt time.Time `json:"UpdatedAt"`
	CreatedAt time.Time `json:"CreatedAt"`

	AnchorAt time.Time `json:"anchor_at"`
	// ScheduleParsed is true when the schedule parsed as a standard 5-field cron.
	ScheduleParsed bool `json:"ScheduleParsed"`
	// ParseError is set when Schedule is non-empty but not a valid standard cron.
	ParseError string `json:"ParseError,omitempty"`
	// ProjectedRuns are RFC3339 (UTC) instants from now through horizon, after
	// anchoring on updated_at. Empty if inactive, parse failed, or no run in range.
	ProjectedRuns []string `json:"ProjectedRuns"`
}

const (
	cronHorizonDefaultHours = 168
	cronMaxRunsDefault      = 64
)

func (s *Server) listCronScheduleTimeline(w http.ResponseWriter, r *http.Request) {
	horizonHours, errH := intQuery(r, "horizon_hours", cronHorizonDefaultHours, 1, 24*365)
	if errH != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid query", errH))
		return
	}
	maxRuns, errM := intQuery(r, "max_runs", cronMaxRunsDefault, 1, 200)
	if errM != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid query", errM))
		return
	}

	var jobs []db.SwarmCronJob
	if err := s.db.Order("name").Find(&jobs).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	if jobs == nil {
		jobs = []db.SwarmCronJob{}
	}

	now := time.Now().UTC()
	horizon := now.Add(time.Duration(horizonHours) * time.Hour)

	rows := make([]CronScheduleTimelineRow, 0, len(jobs))
	for _, j := range jobs {
		row := CronScheduleTimelineRow{
			ID:        j.ID.String(),
			Name:      j.Name,
			AgentName: j.AgentName,
			Schedule:  j.Schedule,
			Active:    j.Active,
			UpdatedAt: j.UpdatedAt,
			CreatedAt: j.CreatedAt,
		}
		anchor := j.UpdatedAt
		if anchor.IsZero() {
			anchor = j.CreatedAt
		}
		row.AnchorAt = anchor
		if !j.Active {
			row.ScheduleParsed = isLikelyStandardCron(j.Schedule)
			rows = append(rows, row)
			continue
		}
		expr := j.Schedule
		sched, err := cron.ParseStandard(expr)
		if err != nil {
			row.ParseError = err.Error()
			row.ProjectedRuns = nil
			rows = append(rows, row)
			continue
		}
		row.ScheduleParsed = true
		row.ProjectedRuns = projectCronRunsForTimeline(sched, anchor, now, horizon, maxRuns)
		rows = append(rows, row)
	}

	out := CronScheduleTimelineResponse{
		AsOf:         now,
		HorizonEnds:  horizon,
		AnchoredBy:   "UpdatedAt",
		HorizonHours: horizonHours,
		MaxRuns:      maxRuns,
		Rows:         rows,
	}
	writeJSON(w, http.StatusOK, out)
}

func isLikelyStandardCron(s string) bool {
	_, err := cron.ParseStandard(s)
	return err == nil
}

// projectCronRunsForTimeline returns upcoming fires between now and horizon (inclusive
// of the window end when it coincides with a run), after anchoring: first tick on/after
// `anchor`, then skip ticks before `asOf` (e.g. missed while paused, or past the minute).
func projectCronRunsForTimeline(
	sched cron.Schedule,
	anchor, asOf, horizon time.Time,
	maxRuns int,
) []string {
	anchor, asOf, horizon = anchor.UTC(), asOf.UTC(), horizon.UTC()
	t := firstCronFireOnOrAfterAnchor(sched, anchor)
	const capAdvance = 1_000_000
	adv := 0
	for t.Before(asOf) && adv < capAdvance {
		t = sched.Next(t)
		adv++
	}
	var out []string
	adv2 := 0
	for len(out) < maxRuns {
		if t.After(horizon) {
			break
		}
		out = append(out, t.Format(time.RFC3339Nano))
		nt := sched.Next(t)
		if nt == t {
			break
		}
		adv2++
		if adv2 > capAdvance {
			break
		}
		t = nt
	}
	return out
}

func firstCronFireOnOrAfterAnchor(sched cron.Schedule, anchor time.Time) time.Time {
	if anchor.IsZero() {
		return time.Now().UTC()
	}
	return sched.Next(anchor.Add(-1 * time.Nanosecond))
}

func intQuery(
	r *http.Request,
	name string, def, min, max int,
) (int, error) {
	qs := r.URL.Query().Get(name)
	if qs == "" {
		return def, nil
	}
	v, err := strconv.Atoi(qs)
	if err != nil {
		return 0, fmt.Errorf("%s: must be an integer", name)
	}
	if v < min || v > max {
		return 0, fmt.Errorf("%s: out of range (%d–%d)", name, min, max)
	}
	return v, nil
}
