package repository

import (
	"context"

	"esports-backend/internal/entity"
)

type AuditRepository struct {
	db Queryer
}

func NewAuditRepository(db Queryer) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO audit_logs (id, actor_user_id, tournament_id, entity_type, entity_id, action_type, description, metadata_json)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
    `, log.ID, log.ActorUserID, log.TournamentID, log.EntityType, log.EntityID, log.ActionType, log.Description, log.MetadataJSON)
	return err
}

func (r *AuditRepository) ListByTournament(ctx context.Context, tournamentID string) ([]entity.AuditLog, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, actor_user_id, tournament_id, entity_type, entity_id, action_type, description, metadata_json, created_at
        FROM audit_logs
        WHERE tournament_id=$1
        ORDER BY created_at DESC
    `, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]entity.AuditLog, 0)
	for rows.Next() {
		var a entity.AuditLog
		if err := rows.Scan(&a.ID, &a.ActorUserID, &a.TournamentID, &a.EntityType, &a.EntityID, &a.ActionType, &a.Description, &a.MetadataJSON, &a.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}
