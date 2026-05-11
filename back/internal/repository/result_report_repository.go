package repository

import (
	"context"
	"errors"

	"esports-backend/internal/entity"

	"github.com/jackc/pgx/v5"
)

type ResultReportRepository struct {
	db Queryer
}

func NewResultReportRepository(db Queryer) *ResultReportRepository {
	return &ResultReportRepository{db: db}
}

func (r *ResultReportRepository) Create(ctx context.Context, rr *entity.ResultReport) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO result_reports (id, match_id, reported_by_id, winner_id, score1, score2, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, rr.ID, rr.MatchID, rr.ReportedByID, rr.WinnerID, rr.Score1, rr.Score2, rr.Status, rr.CreatedAt)
	return err
}

func (r *ResultReportRepository) GetByID(ctx context.Context, id string) (*entity.ResultReport, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, match_id, reported_by_id, winner_id, score1, score2, status, created_at
		FROM result_reports WHERE id=$1
	`, id)
	return scanResultReport(row)
}

func (r *ResultReportRepository) ListPendingByMatch(ctx context.Context, matchID string) ([]*entity.ResultReport, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, match_id, reported_by_id, winner_id, score1, score2, status, created_at
		FROM result_reports WHERE match_id=$1 AND status='pending'
		ORDER BY created_at ASC
	`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanResultReports(rows)
}

func (r *ResultReportRepository) ListByMatch(ctx context.Context, matchID string) ([]*entity.ResultReport, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, match_id, reported_by_id, winner_id, score1, score2, status, created_at
		FROM result_reports WHERE match_id=$1
		ORDER BY created_at DESC
	`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanResultReports(rows)
}

func (r *ResultReportRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE result_reports SET status=$2 WHERE id=$1`, id, status)
	return err
}

func scanResultReport(row interface{ Scan(...interface{}) error }) (*entity.ResultReport, error) {
	var rr entity.ResultReport
	err := row.Scan(&rr.ID, &rr.MatchID, &rr.ReportedByID, &rr.WinnerID, &rr.Score1, &rr.Score2, &rr.Status, &rr.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pgx.ErrNoRows
	}
	return &rr, err
}

func scanResultReports(rows pgx.Rows) ([]*entity.ResultReport, error) {
	var out []*entity.ResultReport
	for rows.Next() {
		rr, err := scanResultReport(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rr)
	}
	return out, rows.Err()
}
