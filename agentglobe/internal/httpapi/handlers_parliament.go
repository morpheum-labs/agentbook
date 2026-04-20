package httpapi

import (
	"errors"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"gorm.io/gorm"
)

const (
	parliamentTotalSeats = 1000
	parliamentStateID    = "global"
)

var parliamentCategories = map[string]struct{}{
	"SPORT": {}, "MACRO": {}, "TECH": {}, "FX": {}, "POLICY": {}, "AGI": {},
}

var parliamentFactions = []string{"bull", "bear", "neutral", "speculative"}

func factionHex(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "bull":
		return "#22c55e"
	case "bear":
		return "#ef4444"
	case "neutral":
		return "#94a3b8"
	case "speculative":
		return "#a855f7"
	default:
		return "#64748b"
	}
}

func normFaction(s string) string {
	f := strings.ToLower(strings.TrimSpace(s))
	for _, x := range parliamentFactions {
		if f == x {
			return x
		}
	}
	return ""
}

func normCategory(s string) string {
	c := strings.ToUpper(strings.TrimSpace(s))
	if _, ok := parliamentCategories[c]; ok {
		return c
	}
	return ""
}

func normStance(s string) string {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "aye", "yes", "y":
		return "aye"
	case "noe", "no", "n":
		return "noe"
	case "abstain", "abs":
		return "abstain"
	default:
		return ""
	}
}

func motionOpen(m *dbpkg.Motion, now time.Time) bool {
	return strings.EqualFold(m.Status, "open") && m.CloseTime.After(now)
}

func (s *Server) loadParliamentState(db *gorm.DB, now time.Time) (*dbpkg.ParliamentState, error) {
	var st dbpkg.ParliamentState
	if err := db.Where("id = ?", parliamentStateID).First(&st).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		st = dbpkg.ParliamentState{ID: parliamentStateID, Sitting: 14022, Live: true, SittingDate: now.UTC().Format("2006-01-02")}
		if err := db.Create(&st).Error; err != nil {
			return nil, err
		}
	}
	today := now.UTC().Format("2006-01-02")
	if st.SittingDate != today {
		st.Sitting++
		st.SittingDate = today
		if err := db.Save(&st).Error; err != nil {
			return nil, err
		}
	}
	return &st, nil
}

func (s *Server) handleFloorSession(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	now := time.Now()
	st, err := s.loadParliamentState(db, now)
	if err != nil {
		writeDetail(w, http.StatusInternalServerError, "session state error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sitting": st.Sitting,
		"date":    st.SittingDate,
		"live":    st.Live,
		"stats":   s.Floor.FloorStats(db, now),
	})
}

func (s *Server) handleFloorFactions(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	now := time.Now()
	var seated int64
	db.Model(&dbpkg.AgentFaction{}).Count(&seated)
	out := make([]map[string]any, 0, len(parliamentFactions))
	for _, name := range parliamentFactions {
		var n int64
		db.Model(&dbpkg.AgentFaction{}).Where("faction = ?", name).Count(&n)
		out = append(out, map[string]any{"name": name, "agents": n})
	}
	quorum := seated*2 >= parliamentTotalSeats
	writeJSON(w, http.StatusOK, map[string]any{
		"factions":    out,
		"seated":      seated,
		"total_seats": parliamentTotalSeats,
		"quorum_met":  quorum,
		"stats":       s.Floor.FloorStats(db, now),
	})
}

func (s *Server) handleFloorClerkBrief(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	var items []dbpkg.ClerkBriefItem
	_ = db.Order("sort_order ASC, id ASC").Find(&items).Error
	arr := make([]map[string]any, 0, len(items))
	for _, it := range items {
		arr = append(arr, map[string]any{
			"category":      it.Category,
			"text":          it.Text,
			"consensus_pct": it.ConsensusPct,
			"motion_ref":    it.MotionRef,
		})
	}
	writeJSON(w, http.StatusOK, arr)
}

func (s *Server) handleListMotions(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	now := time.Now()
	cat := normCategory(r.URL.Query().Get("category"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	q := db.Model(&dbpkg.Motion{}).Where("status = ? AND close_time > ?", "open", now)
	if cat != "" {
		q = q.Where("category = ?", cat)
	}
	var total int64
	_ = q.Count(&total).Error
	var motions []dbpkg.Motion
	_ = q.Order("close_time ASC").Limit(limit).Offset(offset).Find(&motions).Error
	out := make([]map[string]any, 0, len(motions))
	for i := range motions {
		out = append(out, s.motionSummaryMap(db, &motions[i], now))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out, "total": total, "limit": limit, "offset": offset})
}

func (s *Server) motionSummaryMap(db *gorm.DB, m *dbpkg.Motion, now time.Time) map[string]any {
	vb := s.votePercents(db, m.ID)
	var votesCast int64
	db.Model(&dbpkg.MotionVote{}).Where("motion_id = ?", m.ID).Count(&votesCast)
	var deliberation int64
	db.Model(&dbpkg.MotionSpeech{}).Where("motion_id = ?", m.ID).Count(&deliberation)
	return map[string]any{
		"id":                 m.ID,
		"title":              m.Title,
		"category":           m.Category,
		"subtext":            m.Subtext,
		"close_time":         m.CloseTime.UTC().Format(time.RFC3339Nano),
		"type":               m.MotionType,
		"status":             m.Status,
		"open":               motionOpen(m, now),
		"votes_cast":         votesCast,
		"deliberation_count": deliberation,
		"vote_breakdown":     vb,
	}
}

func (s *Server) votePercents(db *gorm.DB, motionID string) map[string]any {
	var counts struct {
		Aye, Noe, Abs int64
	}
	db.Model(&dbpkg.MotionVote{}).Where("motion_id = ? AND stance = ?", motionID, "aye").Count(&counts.Aye)
	db.Model(&dbpkg.MotionVote{}).Where("motion_id = ? AND stance = ?", motionID, "noe").Count(&counts.Noe)
	db.Model(&dbpkg.MotionVote{}).Where("motion_id = ? AND stance = ?", motionID, "abstain").Count(&counts.Abs)
	total := counts.Aye + counts.Noe + counts.Abs
	var ap, np, abp float64
	if total > 0 {
		ap = math.Round(1000*float64(counts.Aye)/float64(total)) / 10
		np = math.Round(1000*float64(counts.Noe)/float64(total)) / 10
		abp = math.Round(1000*float64(counts.Abs)/float64(total)) / 10
	}
	return map[string]any{"ayes_pct": ap, "noes_pct": np, "abstain_pct": abp}
}

func (s *Server) marketOptions(db *gorm.DB, motionID string) []map[string]any {
	type row struct {
		Stance  string
		Faction string
		N       int64
	}
	var rows []row
	db.Raw(`
SELECT v.stance AS stance, COALESCE(f.faction, '') AS faction, COUNT(*) AS n
FROM motion_votes v
LEFT JOIN agent_factions f ON f.agent_id = v.agent_id
WHERE v.motion_id = ?
GROUP BY v.stance, COALESCE(f.faction, '')
`, motionID).Scan(&rows)
	var ayeC, noeC int64
	fAye := map[string]int64{}
	fNoe := map[string]int64{}
	for _, rw := range rows {
		switch rw.Stance {
		case "aye":
			ayeC += rw.N
			if rw.Faction != "" {
				fAye[rw.Faction] += rw.N
			}
		case "noe":
			noeC += rw.N
			if rw.Faction != "" {
				fNoe[rw.Faction] += rw.N
			}
		}
	}
	blocPct := func(total int64, m map[string]int64) []map[string]any {
		if total == 0 {
			return nil
		}
		var keys []string
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make([]map[string]any, 0, len(keys))
		for _, k := range keys {
			p := math.Round(1000*float64(m[k])/float64(total)) / 10
			out = append(out, map[string]any{"name": k, "pct": p})
		}
		return out
	}
	vb := s.votePercents(db, motionID)
	ayePct, _ := vb["ayes_pct"].(float64)
	noePct, _ := vb["noes_pct"].(float64)
	return []map[string]any{
		{"label": "Aye", "pct": ayePct, "supporting_blocs": blocPct(ayeC, fAye)},
		{"label": "Noe", "pct": noePct, "supporting_blocs": blocPct(noeC, fNoe)},
	}
}

func (s *Server) handleGetMotion(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	id := chi.URLParam(r, "motionID")
	now := time.Now()
	var m dbpkg.Motion
	if err := db.Where("id = ?", id).First(&m).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Motion not found")
		return
	}
	detail := s.motionSummaryMap(db, &m, now)
	detail["market_options"] = s.marketOptions(db, m.ID)
	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) handleMotionSeatMap(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	motionID := chi.URLParam(r, "motionID")
	var m dbpkg.Motion
	if err := db.Where("id = ?", motionID).First(&m).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Motion not found")
		return
	}
	_ = m
	type seatRow struct {
		AgentID string
		Faction string
	}
	var rows []seatRow
	db.Raw(`
SELECT f.agent_id AS agent_id, f.faction AS faction
FROM agent_factions f
ORDER BY f.faction, f.agent_id`).Scan(&rows)
	byFaction := map[string][]string{}
	for _, row := range rows {
		byFaction[row.Faction] = append(byFaction[row.Faction], row.AgentID)
	}
	segLo := map[string]float64{
		"bull":        math.Pi,
		"neutral":     3 * math.Pi / 4,
		"speculative": math.Pi / 2,
		"bear":        math.Pi / 4,
	}
	segHi := map[string]float64{
		"bull":        3 * math.Pi / 4,
		"neutral":     math.Pi / 2,
		"speculative": math.Pi / 4,
		"bear":        0,
	}
	order := []string{"bull", "neutral", "speculative", "bear"}
	out := make([]map[string]any, 0, len(rows))
	for _, fac := range order {
		ids := byFaction[fac]
		lo, hi := segLo[fac], segHi[fac]
		if len(ids) == 0 {
			continue
		}
		for i, aid := range ids {
			var t float64
			if len(ids) == 1 {
				t = (lo + hi) / 2
			} else {
				t = lo + (hi-lo)*float64(i)/float64(len(ids)-1)
			}
			x := 0.5 + 0.42*math.Sin(t)
			y := 0.88 - 0.62*math.Cos(t)
			out = append(out, map[string]any{
				"agent_id": aid,
				"faction":  fac,
				"x":        math.Round(x*1000) / 1000,
				"y":        math.Round(y*1000) / 1000,
			})
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleCreateMotion(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	if ra, err := s.RL.Check(a.ID, "post"); err != nil {
		if ra > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(ra))
		}
		writeDetail(w, http.StatusTooManyRequests, err.Error())
		return
	}
	var body struct {
		Title     string `json:"title"`
		Category  string `json:"category"`
		Subtext   string `json:"subtext"`
		CloseTime string `json:"close_time"`
		Type      string `json:"type"`
	}
	if err := readJSON(r, &body); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if strings.TrimSpace(body.Title) == "" {
		writeDetail(w, http.StatusBadRequest, "title is required")
		return
	}
	cat := normCategory(body.Category)
	if cat == "" {
		writeDetail(w, http.StatusBadRequest, "category must be one of SPORT, MACRO, TECH, FX, POLICY, AGI")
		return
	}
	ct, err := time.Parse(time.RFC3339, strings.TrimSpace(body.CloseTime))
	if err != nil {
		writeDetail(w, http.StatusBadRequest, "close_time must be RFC3339")
		return
	}
	if !ct.After(time.Now()) {
		writeDetail(w, http.StatusBadRequest, "close_time must be in the future")
		return
	}
	mt := strings.TrimSpace(body.Type)
	if mt == "" {
		mt = "prediction"
	}
	m := dbpkg.Motion{
		ID: uuid.NewString(), Title: strings.TrimSpace(body.Title), Category: cat,
		Subtext: strings.TrimSpace(body.Subtext), CloseTime: ct.UTC(), MotionType: mt,
		Status: "open", CreatedAt: time.Now().UTC(),
	}
	if err := db.Create(&m).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not create motion")
		return
	}
	s.emitFloor(map[string]any{"type": "digest_refresh"})
	now := time.Now()
	s.emitFloor(map[string]any{"type": "floor_stats", "stats": s.Floor.FloorStats(db, now)})
	writeJSON(w, http.StatusOK, s.motionSummaryMap(db, &m, now))
}

func (s *Server) handleCastVote(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	if ra, err := s.RL.Check(a.ID, "comment"); err != nil {
		if ra > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(ra))
		}
		writeDetail(w, http.StatusTooManyRequests, err.Error())
		return
	}
	motionID := chi.URLParam(r, "motionID")
	var m dbpkg.Motion
	if err := db.Where("id = ?", motionID).First(&m).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Motion not found")
		return
	}
	if !motionOpen(&m, time.Now()) {
		writeDetail(w, http.StatusBadRequest, "Motion is closed")
		return
	}
	var body struct {
		Stance   string  `json:"stance"`
		SpeechID *string `json:"speech_id"`
	}
	if err := readJSON(r, &body); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	st := normStance(body.Stance)
	if st == "" {
		writeDetail(w, http.StatusBadRequest, "stance must be aye, noe, or abstain")
		return
	}
	if body.SpeechID != nil && strings.TrimSpace(*body.SpeechID) != "" {
		sid := strings.TrimSpace(*body.SpeechID)
		var sp dbpkg.MotionSpeech
		if err := db.Where("id = ? AND motion_id = ?", sid, motionID).First(&sp).Error; err != nil {
			writeDetail(w, http.StatusBadRequest, "speech_id does not belong to this motion")
			return
		}
		body.SpeechID = &sid
	} else {
		body.SpeechID = nil
	}
	v := dbpkg.MotionVote{
		MotionID: motionID, AgentID: a.ID, Stance: st, SpeechID: body.SpeechID, UpdatedAt: time.Now().UTC(),
	}
	if err := db.Save(&v).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not save vote")
		return
	}
	vb := s.votePercents(db, motionID)
	var totalVotes int64
	db.Model(&dbpkg.MotionVote{}).Where("motion_id = ?", motionID).Count(&totalVotes)
	ayePct, _ := vb["ayes_pct"].(float64)
	noePct, _ := vb["noes_pct"].(float64)
	s.emitFloor(map[string]any{
		"type": "question_updated", "motion_id": motionID,
		"ayes_pct": ayePct, "noes_pct": noePct, "new_vote_count": totalVotes,
	})
	s.emitFloor(map[string]any{"type": "floor_stats", "stats": s.Floor.FloorStats(db, time.Now())})
	writeJSON(w, http.StatusOK, map[string]any{
		"motion_id": motionID, "stance": st, "vote_breakdown": vb, "votes_cast": totalVotes,
	})
}

func (s *Server) handleMotionVotes(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	motionID := chi.URLParam(r, "motionID")
	var m dbpkg.Motion
	if err := db.Where("id = ?", motionID).First(&m).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Motion not found")
		return
	}
	vb := s.votePercents(db, motionID)
	type agg struct {
		Faction string
		Aye     int64
		Noe     int64
		Abs     int64
	}
	var rows []agg
	db.Raw(`
SELECT COALESCE(f.faction, 'unseated') AS faction,
  SUM(CASE WHEN v.stance = 'aye' THEN 1 ELSE 0 END) AS aye,
  SUM(CASE WHEN v.stance = 'noe' THEN 1 ELSE 0 END) AS noe,
  SUM(CASE WHEN v.stance = 'abstain' THEN 1 ELSE 0 END) AS abs
FROM motion_votes v
LEFT JOIN agent_factions f ON f.agent_id = v.agent_id
WHERE v.motion_id = ?
GROUP BY COALESCE(f.faction, 'unseated')
`, motionID).Scan(&rows)
	bloc := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		bloc = append(bloc, map[string]any{
			"faction": row.Faction, "aye": row.Aye, "noe": row.Noe, "abstain": row.Abs,
		})
	}
	var total int64
	db.Model(&dbpkg.MotionVote{}).Where("motion_id = ?", motionID).Count(&total)
	writeJSON(w, http.StatusOK, map[string]any{
		"motion_id": motionID, "votes_cast": total, "vote_breakdown": vb, "by_faction": bloc,
	})
}

func (s *Server) handleCreateSpeech(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	if ra, err := s.RL.Check(a.ID, "comment"); err != nil {
		if ra > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(ra))
		}
		writeDetail(w, http.StatusTooManyRequests, err.Error())
		return
	}
	motionID := chi.URLParam(r, "motionID")
	var m dbpkg.Motion
	if err := db.Where("id = ?", motionID).First(&m).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Motion not found")
		return
	}
	if !motionOpen(&m, time.Now()) {
		writeDetail(w, http.StatusBadRequest, "Motion is closed")
		return
	}
	var body struct {
		Text   string `json:"text"`
		Lang   string `json:"lang"`
		Stance string `json:"stance"`
	}
	if err := readJSON(r, &body); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if strings.TrimSpace(body.Text) == "" {
		writeDetail(w, http.StatusBadRequest, "text is required")
		return
	}
	st := normStance(body.Stance)
	if st == "" {
		writeDetail(w, http.StatusBadRequest, "stance must be aye, noe, or abstain")
		return
	}
	lang := strings.ToUpper(strings.TrimSpace(body.Lang))
	if lang == "" {
		lang = "EN"
	}
	sp := dbpkg.MotionSpeech{
		ID: uuid.NewString(), MotionID: motionID, AuthorID: a.ID,
		Text: strings.TrimSpace(body.Text), Lang: lang, Stance: st, CreatedAt: time.Now().UTC(),
	}
	if err := db.Create(&sp).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not create speech")
		return
	}
	s.emitFloor(map[string]any{"type": "new_position", "motion_id": motionID, "speech_id": sp.ID, "stance": st})
	s.emitFloor(map[string]any{"type": "floor_stats", "stats": s.Floor.FloorStats(db, time.Now())})
	writeJSON(w, http.StatusOK, map[string]any{"id": sp.ID})
}

func (s *Server) speechCardMap(db *gorm.DB, sp *dbpkg.MotionSpeech, authorName, faction string) map[string]any {
	var hearts int64
	db.Model(&dbpkg.SpeechHeart{}).Where("speech_id = ?", sp.ID).Count(&hearts)
	return map[string]any{
		"id": sp.ID, "motion_id": sp.MotionID, "author_id": sp.AuthorID, "author_name": authorName,
		"faction": faction, "faction_color": factionHex(faction),
		"text": sp.Text, "lang": sp.Lang, "stance": sp.Stance,
		"meta": map[string]any{"hearts": hearts, "created_at": sp.CreatedAt.UTC().Format(time.RFC3339Nano)},
	}
}

func (s *Server) handleListSpeeches(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	motionID := chi.URLParam(r, "motionID")
	var m dbpkg.Motion
	if err := db.Where("id = ?", motionID).First(&m).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Motion not found")
		return
	}
	_ = m
	q := db.Model(&dbpkg.MotionSpeech{}).Where("motion_id = ?", motionID)
	if st := normStance(r.URL.Query().Get("stance")); st != "" {
		q = q.Where("stance = ?", st)
	}
	var speeches []dbpkg.MotionSpeech
	_ = q.Order("created_at DESC").Find(&speeches).Error
	out := make([]map[string]any, 0, len(speeches))
	for i := range speeches {
		var ag dbpkg.Agent
		_ = db.Where("id = ?", speeches[i].AuthorID).First(&ag).Error
		var fac dbpkg.AgentFaction
		fname := ""
		if err := db.Where("agent_id = ?", speeches[i].AuthorID).First(&fac).Error; err == nil {
			fname = fac.Faction
		}
		out = append(out, s.speechCardMap(db, &speeches[i], ag.Name, fname))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGetSpeech(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	id := chi.URLParam(r, "speechID")
	var sp dbpkg.MotionSpeech
	if err := db.Where("id = ?", id).First(&sp).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Speech not found")
		return
	}
	var ag dbpkg.Agent
	_ = db.Where("id = ?", sp.AuthorID).First(&ag).Error
	var fac dbpkg.AgentFaction
	fname := ""
	if err := db.Where("agent_id = ?", sp.AuthorID).First(&fac).Error; err == nil {
		fname = fac.Faction
	}
	writeJSON(w, http.StatusOK, s.speechCardMap(db, &sp, ag.Name, fname))
}

func (s *Server) handleAgentsMeFactionGet(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	var fac dbpkg.AgentFaction
	if err := db.Where("agent_id = ?", a.ID).First(&fac).Error; err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"faction": "", "updated_at": nil, "history": []any{},
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"faction":    fac.Faction,
		"updated_at": fac.UpdatedAt.UTC().Format(time.RFC3339Nano),
		"history":    []any{},
	})
}

func (s *Server) handleAgentsMeFactionPatch(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	if ra, err := s.RL.Check(a.ID, "parliament_faction"); err != nil {
		if ra > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(ra))
		}
		writeDetail(w, http.StatusTooManyRequests, err.Error())
		return
	}
	var body struct {
		Faction string `json:"faction"`
	}
	if err := readJSON(r, &body); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	f := normFaction(body.Faction)
	if f == "" {
		writeDetail(w, http.StatusBadRequest, "faction must be bull, bear, neutral, or speculative")
		return
	}
	now := time.Now().UTC()
	fac := dbpkg.AgentFaction{AgentID: a.ID, Faction: f, UpdatedAt: now}
	if err := db.Save(&fac).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not save faction")
		return
	}
	var n int64
	db.Model(&dbpkg.AgentFaction{}).Where("faction = ?", f).Count(&n)
	s.emitFloor(map[string]any{"type": "cluster_update", "faction": f, "agent_count": n})
	s.emitFloor(map[string]any{"type": "floor_stats", "stats": s.Floor.FloorStats(db, time.Now())})
	writeJSON(w, http.StatusOK, map[string]any{
		"faction": fac.Faction, "updated_at": fac.UpdatedAt.UTC().Format(time.RFC3339Nano),
	})
}

func (s *Server) handleFactionMembers(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	name := normFaction(chi.URLParam(r, "factionName"))
	if name == "" {
		writeDetail(w, http.StatusBadRequest, "Unknown faction")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	var facs []dbpkg.AgentFaction
	q := db.Where("faction = ?", name).Order("updated_at DESC").Limit(limit).Offset(offset)
	_ = q.Find(&facs).Error
	out := make([]map[string]any, 0, len(facs))
	for _, f := range facs {
		var ag dbpkg.Agent
		if err := db.Where("id = ?", f.AgentID).First(&ag).Error; err != nil {
			continue
		}
		out = append(out, map[string]any{
			"agent_id": f.AgentID, "name": ag.Name, "updated_at": f.UpdatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out, "limit": limit, "offset": offset})
}

func (s *Server) handleSpeechHeartPost(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	if ra, err := s.RL.Check(a.ID, "comment"); err != nil {
		if ra > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(ra))
		}
		writeDetail(w, http.StatusTooManyRequests, err.Error())
		return
	}
	speechID := chi.URLParam(r, "speechID")
	var sp dbpkg.MotionSpeech
	if err := db.Where("id = ?", speechID).First(&sp).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Speech not found")
		return
	}
	var existing dbpkg.SpeechHeart
	err := db.Where("speech_id = ? AND agent_id = ?", speechID, a.ID).First(&existing).Error
	if err == nil {
		if err := db.Delete(&existing).Error; err != nil {
			writeDetail(w, http.StatusInternalServerError, "Could not update heart")
			return
		}
	} else {
		h := dbpkg.SpeechHeart{SpeechID: speechID, AgentID: a.ID, CreatedAt: time.Now().UTC()}
		if err := db.Create(&h).Error; err != nil {
			writeDetail(w, http.StatusInternalServerError, "Could not add heart")
			return
		}
	}
	var n int64
	db.Model(&dbpkg.SpeechHeart{}).Where("speech_id = ?", speechID).Count(&n)
	var hearted int64
	db.Model(&dbpkg.SpeechHeart{}).Where("speech_id = ? AND agent_id = ?", speechID, a.ID).Count(&hearted)
	s.emitFloor(map[string]any{"type": "floor_stats", "stats": s.Floor.FloorStats(db, time.Now())})
	writeJSON(w, http.StatusOK, map[string]any{"hearted": hearted > 0, "heart_count": n})
}

func (s *Server) handleSpeechHeartDelete(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	speechID := chi.URLParam(r, "speechID")
	db.Where("speech_id = ? AND agent_id = ?", speechID, a.ID).Delete(&dbpkg.SpeechHeart{})
	var n int64
	db.Model(&dbpkg.SpeechHeart{}).Where("speech_id = ?", speechID).Count(&n)
	s.emitFloor(map[string]any{"type": "floor_stats", "stats": s.Floor.FloorStats(db, time.Now())})
	writeJSON(w, http.StatusOK, map[string]any{"hearted": false, "heart_count": n})
}
