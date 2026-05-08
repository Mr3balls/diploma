package repository

import (
	"context"
	"errors"

	"esports-backend/internal/entity"

	"github.com/jackc/pgx/v5"
)

type ImportRepository struct {
	db Queryer
}

func NewImportRepository(db Queryer) *ImportRepository {
	return &ImportRepository{db: db}
}

func (r *ImportRepository) UpsertSheetLink(ctx context.Context, s *entity.GoogleSheetLink) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO google_sheet_links (id, tournament_id, sheet_url, spreadsheet_id, worksheet_name, status, created_by)
        VALUES ($1,$2,$3,$4,$5,$6,$7)
        ON CONFLICT (tournament_id)
        DO UPDATE SET sheet_url=excluded.sheet_url, spreadsheet_id=excluded.spreadsheet_id, worksheet_name=excluded.worksheet_name, status=excluded.status, updated_at=now()
    `, s.ID, s.TournamentID, s.SheetURL, s.SpreadsheetID, s.WorksheetName, s.Status, s.CreatedBy)
	return err
}

func (r *ImportRepository) GetSheetLinkByTournamentID(ctx context.Context, tournamentID string) (*entity.GoogleSheetLink, error) {
	row := r.db.QueryRow(ctx, `
        SELECT id, tournament_id, sheet_url, spreadsheet_id, worksheet_name, status, created_by, created_at, updated_at
        FROM google_sheet_links
        WHERE tournament_id=$1
    `, tournamentID)
	var s entity.GoogleSheetLink
	err := row.Scan(&s.ID, &s.TournamentID, &s.SheetURL, &s.SpreadsheetID, &s.WorksheetName, &s.Status, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &s, nil
}

func (r *ImportRepository) CreateBatch(ctx context.Context, b *entity.ImportBatch) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO import_batches (id, tournament_id, sheet_link_id, started_by, status, summary_json)
        VALUES ($1,$2,$3,$4,$5,$6)
    `, b.ID, b.TournamentID, b.SheetLinkID, b.StartedBy, b.Status, b.SummaryJSON)
	return err
}

func (r *ImportRepository) UpdateBatch(ctx context.Context, b *entity.ImportBatch) error {
	_, err := r.db.Exec(ctx, `UPDATE import_batches SET status=$2, summary_json=$3, updated_at=now() WHERE id=$1`, b.ID, b.Status, b.SummaryJSON)
	return err
}

func (r *ImportRepository) GetBatch(ctx context.Context, batchID string) (*entity.ImportBatch, error) {
	row := r.db.QueryRow(ctx, `
        SELECT id, tournament_id, sheet_link_id, started_by, status, summary_json, created_at, updated_at
        FROM import_batches
        WHERE id=$1
    `, batchID)
	var b entity.ImportBatch
	err := row.Scan(&b.ID, &b.TournamentID, &b.SheetLinkID, &b.StartedBy, &b.Status, &b.SummaryJSON, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &b, nil
}

func (r *ImportRepository) ListBatchesByTournament(ctx context.Context, tournamentID string) ([]entity.ImportBatch, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, tournament_id, sheet_link_id, started_by, status, summary_json, created_at, updated_at
        FROM import_batches
        WHERE tournament_id=$1
        ORDER BY created_at DESC
    `, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.ImportBatch, 0)
	for rows.Next() {
		var b entity.ImportBatch
		if err := rows.Scan(&b.ID, &b.TournamentID, &b.SheetLinkID, &b.StartedBy, &b.Status, &b.SummaryJSON, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func (r *ImportRepository) CreateRow(ctx context.Context, row *entity.ImportRow) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO import_rows (id, batch_id, row_number, raw_data_json, team_name, discipline, captain_nick, player_2_nick, player_3_nick, player_4_nick, player_5_nick, substitute_nick, status, validation_errors_json)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
    `, row.ID, row.BatchID, row.RowNumber, row.RawDataJSON, row.TeamName, row.Discipline, row.CaptainNick, row.Player2Nick, row.Player3Nick, row.Player4Nick, row.Player5Nick, row.SubstituteNick, row.Status, row.ValidationErrorsJSON)
	return err
}

func (r *ImportRepository) UpdateRowStatus(ctx context.Context, id, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE import_rows SET status=$2, updated_at=now() WHERE id=$1`, id, status)
	return err
}

func (r *ImportRepository) ListRowsByBatch(ctx context.Context, batchID string) ([]entity.ImportRow, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, batch_id, row_number, raw_data_json, team_name, discipline, captain_nick, player_2_nick, player_3_nick, player_4_nick, player_5_nick, substitute_nick, status, validation_errors_json, created_at, updated_at
        FROM import_rows
        WHERE batch_id=$1
        ORDER BY row_number ASC
    `, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.ImportRow, 0)
	for rows.Next() {
		var row entity.ImportRow
		if err := rows.Scan(&row.ID, &row.BatchID, &row.RowNumber, &row.RawDataJSON, &row.TeamName, &row.Discipline, &row.CaptainNick, &row.Player2Nick, &row.Player3Nick, &row.Player4Nick, &row.Player5Nick, &row.SubstituteNick, &row.Status, &row.ValidationErrorsJSON, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}
