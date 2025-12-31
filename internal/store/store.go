package store

import "swarmtest/internal/models"

// MissionStore interface
type MissionStore interface {
	Put(mission *models.Mission)
	PutAgent(agent *models.Agent)
	Get(id string) (*models.Mission, bool)
	List() []*models.Mission
	AddActionLog(log models.ActionLog, missionID string)
}
