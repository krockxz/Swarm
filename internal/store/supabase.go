package store

import (
	"database/sql"
	"log"
	
	"swarmtest/internal/models"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// SupabaseStore implements the Store interface using Supabase Postgres
type SupabaseStore struct {
	db *sql.DB
}

func NewSupabaseStore(db *sql.DB) *SupabaseStore {
	return &SupabaseStore{db: db}
}

func (s *SupabaseStore) Put(mission *models.Mission) {
	query := `
		INSERT INTO missions (
			id, name, target_url, num_agents, goal, max_duration_seconds, 
			rate_limit_per_second, initial_system_prompt, status, created_at, 
			started_at, completed_at, total_actions, total_errors, 
			average_latency_ms, completed_agents, failed_agents
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			started_at = EXCLUDED.started_at,
			completed_at = EXCLUDED.completed_at,
			total_actions = EXCLUDED.total_actions,
			total_errors = EXCLUDED.total_errors,
			average_latency_ms = EXCLUDED.average_latency_ms,
			completed_agents = EXCLUDED.completed_agents,
			failed_agents = EXCLUDED.failed_agents;
	`

	_, err := s.db.Exec(query,
		mission.ID, mission.Name, mission.TargetURL, mission.NumAgents, mission.Goal,
		mission.MaxDurationSeconds, mission.RateLimitPerSecond, mission.InitialSystemPrompt,
		mission.Status, mission.CreatedAt, mission.StartedAt, mission.CompletedAt,
		mission.TotalActions, mission.TotalErrors, mission.AverageLatencyMS,
		mission.CompletedAgents, mission.FailedAgents,
	)
	if err != nil {
		log.Printf("Error saving mission %s: %v", mission.ID, err)
		return
	}

	// Update agents
	for _, agent := range mission.AgentMetrics {
		s.PutAgent(agent)
	}
	
	// Note: RecentEvents are not bulk updated here to avoid perf issues. 
	// The event logger should handle them, or we assume they are inserted via AddActionLog
}

func (s *SupabaseStore) PutAgent(agent *models.Agent) {
	query := `
		INSERT INTO agents (
			id, mission_id, status, current_url, error_count, success_count,
			total_latency_ms, consecutive_errors, last_action_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			current_url = EXCLUDED.current_url,
			error_count = EXCLUDED.error_count,
			success_count = EXCLUDED.success_count,
			total_latency_ms = EXCLUDED.total_latency_ms,
			consecutive_errors = EXCLUDED.consecutive_errors,
			last_action_at = EXCLUDED.last_action_at;
	`
	_, err := s.db.Exec(query,
		agent.ID, agent.MissionID, agent.Status, agent.CurrentURL,
		agent.ErrorCount, agent.SuccessCount, agent.TotalLatencyMS,
		agent.ConsecutiveErrors, agent.LastActionAt,
	)
	if err != nil {
		log.Printf("Error saving agent %s: %v", agent.ID, err)
	}
}

func (s *SupabaseStore) Get(id string) (*models.Mission, bool) {
	m := &models.Mission{}
	
	// Get Mission
	query := `
		SELECT id, name, target_url, num_agents, goal, max_duration_seconds,
		       rate_limit_per_second, initial_system_prompt, status, created_at,
		       started_at, completed_at, total_actions, total_errors,
		       average_latency_ms, completed_agents, failed_agents
		FROM missions WHERE id = $1`
		
	err := s.db.QueryRow(query, id).Scan(
		&m.ID, &m.Name, &m.TargetURL, &m.NumAgents, &m.Goal, &m.MaxDurationSeconds,
		&m.RateLimitPerSecond, &m.InitialSystemPrompt, &m.Status, &m.CreatedAt,
		&m.StartedAt, &m.CompletedAt, &m.TotalActions, &m.TotalErrors,
		&m.AverageLatencyMS, &m.CompletedAgents, &m.FailedAgents,
	)
	if err == sql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		log.Printf("Error getting mission %s: %v", id, err)
		return nil, false
	}

	// Get Agents
	m.AgentMetrics = make(map[string]*models.Agent)
	agentQuery := `SELECT id, mission_id, status, current_url, error_count, success_count, total_latency_ms, consecutive_errors, last_action_at FROM agents WHERE mission_id = $1`
	rows, err := s.db.Query(agentQuery, id)
	if err != nil {
		log.Printf("Error getting agents for mission %s: %v", id, err)
	} else {
		defer rows.Close()
		for rows.Next() {
			a := &models.Agent{}
			if err := rows.Scan(
				&a.ID, &a.MissionID, &a.Status, &a.CurrentURL, &a.ErrorCount,
				&a.SuccessCount, &a.TotalLatencyMS, &a.ConsecutiveErrors, &a.LastActionAt,
			); err != nil {
				continue
			}
			m.AgentMetrics[a.ID] = a
		}
	}
	
	// Get Recent Events (Logs)
	// We'll just get the last 20 logs
	logQuery := `
		SELECT timestamp, agent_id, action, selector, result, latency_ms, error_message, new_url
		FROM action_logs
		WHERE mission_id = $1
		ORDER BY id DESC
		LIMIT 20`
		
	logRows, err := s.db.Query(logQuery, id)
	if err != nil {
		log.Printf("Error getting logs for mission %s: %v", id, err)
	} else {
		defer logRows.Close()
		for logRows.Next() {
			l := models.ActionLog{}
			var selector, errMsg, newUrl sql.NullString
			if err := logRows.Scan(
				&l.Timestamp, &l.AgentID, &l.Action, &selector, &l.Result,
				&l.LatencyMS, &errMsg, &newUrl,
			); err != nil {
				continue
			}
			l.Selector = selector.String
			l.ErrorMessage = errMsg.String
			l.NewURL = newUrl.String
			
			m.RecentEvents = append(m.RecentEvents, l)
		}
	}

	return m, true
}

func (s *SupabaseStore) List() []*models.Mission {
	query := `
		SELECT id, name, target_url, num_agents, goal, max_duration_seconds,
		       rate_limit_per_second, initial_system_prompt, status, created_at,
		       started_at, completed_at, total_actions, total_errors,
		       average_latency_ms, completed_agents, failed_agents
		FROM missions ORDER BY created_at DESC LIMIT 50`
		
	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("Error listing missions: %v", err)
		return []*models.Mission{}
	}
	defer rows.Close()
	
	var missions []*models.Mission
	for rows.Next() {
		m := &models.Mission{}
		if err := rows.Scan(
			&m.ID, &m.Name, &m.TargetURL, &m.NumAgents, &m.Goal, &m.MaxDurationSeconds,
			&m.RateLimitPerSecond, &m.InitialSystemPrompt, &m.Status, &m.CreatedAt,
			&m.StartedAt, &m.CompletedAt, &m.TotalActions, &m.TotalErrors,
			&m.AverageLatencyMS, &m.CompletedAgents, &m.FailedAgents,
		); err != nil {
			continue
		}
		// Note: We don't populate Agents/Logs for list view to keep it fast
		missions = append(missions, m)
	}
	return missions
}

func (s *SupabaseStore) AddActionLog(logEntry models.ActionLog, missionID string) {
	query := `
		INSERT INTO action_logs (
			timestamp, mission_id, agent_id, action, selector, result, 
			latency_ms, error_message, new_url
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	
	_, err := s.db.Exec(query,
		logEntry.Timestamp, missionID, logEntry.AgentID, logEntry.Action,
		ToNullString(logEntry.Selector), logEntry.Result, logEntry.LatencyMS,
		ToNullString(logEntry.ErrorMessage), ToNullString(logEntry.NewURL),
	)
	if err != nil {
		log.Printf("Error adding log: %v", err)
	}
}

func ToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
