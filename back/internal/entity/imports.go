package entity

import (
	"encoding/json"
	"time"
)

type GoogleSheetLink struct {
	ID            string    `json:"id"`
	TournamentID  string    `json:"tournament_id"`
	SheetURL      string    `json:"sheet_url"`
	SpreadsheetID string    `json:"spreadsheet_id"`
	WorksheetName string    `json:"worksheet_name"`
	Status        string    `json:"status"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ImportBatch struct {
	ID           string          `json:"id"`
	TournamentID string          `json:"tournament_id"`
	SheetLinkID  string          `json:"sheet_link_id"`
	StartedBy    string          `json:"started_by"`
	Status       string          `json:"status"`
	SummaryJSON  json.RawMessage `json:"summary_json"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type ImportRow struct {
	ID                   string          `json:"id"`
	BatchID              string          `json:"batch_id"`
	RowNumber            int             `json:"row_number"`
	RawDataJSON          json.RawMessage `json:"raw_data_json"`
	TeamName             string          `json:"team_name"`
	Discipline           string          `json:"discipline"`
	CaptainNick          string          `json:"captain_nick"`
	Player2Nick          string          `json:"player_2_nick"`
	Player3Nick          string          `json:"player_3_nick"`
	Player4Nick          string          `json:"player_4_nick"`
	Player5Nick          string          `json:"player_5_nick"`
	SubstituteNick       string          `json:"substitute_nick"`
	Status               string          `json:"status"`
	ValidationErrorsJSON json.RawMessage `json:"validation_errors_json"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}
