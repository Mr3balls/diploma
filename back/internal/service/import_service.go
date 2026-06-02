package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"esports-backend/internal/apperror"
	"esports-backend/internal/entity"
	"esports-backend/internal/pkg/notif"
	"esports-backend/internal/pkg/xjson"
	"esports-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SheetsReader interface {
	ReadRows(ctx context.Context, sheetURL, worksheetName string) ([][]string, error)
}

type ImportService struct {
	tournaments   *TournamentService
	imports       *repository.ImportRepository
	teams         *repository.TeamRepository
	users         *repository.UserRepository
	notifications *repository.NotificationRepository
	audits        *repository.AuditRepository
	sheets        SheetsReader
	email         *EmailService
}

func NewImportService(tournaments *TournamentService, imports *repository.ImportRepository, teams *repository.TeamRepository, users *repository.UserRepository, notifications *repository.NotificationRepository, audits *repository.AuditRepository, sheets SheetsReader, email *EmailService) *ImportService {
	return &ImportService{tournaments: tournaments, imports: imports, teams: teams, users: users, notifications: notifications, audits: audits, sheets: sheets, email: email}
}

type ConnectSheetInput struct {
	SheetURL      string
	WorksheetName string
}

type PreviewSummary struct {
	TotalRows        int      `json:"total_rows"`
	ValidRows        int      `json:"valid_rows"`
	NeedsReviewRows  int      `json:"needs_review_rows"`
	DuplicateRows    int      `json:"duplicate_rows"`
	RejectedRows     int      `json:"rejected_rows"`
	DuplicateNicks   []string `json:"duplicate_nicks"`
	PreviewCreatedAt string   `json:"preview_created_at"`
}

func (s *ImportService) Connect(ctx context.Context, actorUserID, tournamentID string, in ConnectSheetInput) (*entity.GoogleSheetLink, error) {
	ok, err := s.tournaments.CanManageTournament(ctx, tournamentID, actorUserID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, apperror.Forbidden("insufficient tournament permissions")
	}
	spreadsheetID, err := parseSpreadsheetID(in.SheetURL)
	if err != nil {
		return nil, apperror.BadRequest("invalid_sheet_url", err.Error(), nil)
	}
	if in.WorksheetName == "" {
		in.WorksheetName = "Sheet1"
	}
	link := &entity.GoogleSheetLink{
		ID:            uuid.NewString(),
		TournamentID:  tournamentID,
		SheetURL:      in.SheetURL,
		SpreadsheetID: spreadsheetID,
		WorksheetName: in.WorksheetName,
		Status:        "connected",
		CreatedBy:     actorUserID,
	}
	if err := s.imports.UpsertSheetLink(ctx, link); err != nil {
		return nil, err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &tournamentID, EntityType: "google_sheet_link", EntityID: link.ID, ActionType: "sheet_connected", Description: "Google Sheet connected", MetadataJSON: xjson.MustMarshal(link)})
	return s.imports.GetSheetLinkByTournamentID(ctx, tournamentID)
}

func (s *ImportService) Validate(ctx context.Context, actorUserID, tournamentID string) (map[string]interface{}, error) {
	ok, err := s.tournaments.CanManageTournament(ctx, tournamentID, actorUserID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, apperror.Forbidden("insufficient tournament permissions")
	}
	link, err := s.imports.GetSheetLinkByTournamentID(ctx, tournamentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("google sheet link not found")
		}
		return nil, err
	}
	rows, err := s.sheets.ReadRows(ctx, link.SheetURL, link.WorksheetName)
	if err != nil {
		return nil, err
	}
	sample := []string{}
	if len(rows) > 0 {
		sample = rows[0]
	}
	return map[string]interface{}{
		"status":         "ok",
		"spreadsheet_id": link.SpreadsheetID,
		"worksheet_name": link.WorksheetName,
		"row_count":      len(rows),
		"sample_row":     sample,
	}, nil
}

func (s *ImportService) Preview(ctx context.Context, actorUserID, tournamentID string) (*entity.ImportBatch, []entity.ImportRow, error) {
	ok, err := s.tournaments.CanManageTournament(ctx, tournamentID, actorUserID)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, apperror.Forbidden("insufficient tournament permissions")
	}
	link, err := s.imports.GetSheetLinkByTournamentID(ctx, tournamentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, apperror.NotFound("google sheet link not found")
		}
		return nil, nil, err
	}
	rows, err := s.sheets.ReadRows(ctx, link.SheetURL, link.WorksheetName)
	if err != nil {
		return nil, nil, err
	}
	batch := &entity.ImportBatch{
		ID:           uuid.NewString(),
		TournamentID: tournamentID,
		SheetLinkID:  link.ID,
		StartedBy:    actorUserID,
		Status:       entity.ImportBatchStatusParsing,
		SummaryJSON:  xjson.MustMarshal(map[string]interface{}{"status": "parsing"}),
	}
	if err := s.imports.CreateBatch(ctx, batch); err != nil {
		return nil, nil, err
	}

	duplicateEmailOwners := make(map[string]int)
	parsedRows := make([]entity.ImportRow, 0)
	hasHeader := false
	if len(rows) > 0 {
		first := normalizeRow(rows[0])
		if len(first) > 0 && strings.Contains(strings.ToLower(strings.Join(first, ",")), "team") {
			hasHeader = true
		}
	}

	startIndex := 0
	if hasHeader {
		startIndex = 1
	}

	summary := PreviewSummary{PreviewCreatedAt: time.Now().Format(time.RFC3339)}
	for idx := startIndex; idx < len(rows); idx++ {
		rowCells := normalizeRow(rows[idx])
		record := toImportRow(batch.ID, idx+1, rowCells)
		errs := validateImportRow(record)
		for _, email := range nonEmptyMembers(record) {
			duplicateEmailOwners[strings.ToLower(email)]++
		}
		record.ValidationErrorsJSON = xjson.MustMarshal(errs)
		if len(errs) == 0 {
			record.Status = entity.ImportRowStatusValid
			summary.ValidRows++
		} else {
			record.Status = entity.ImportRowStatusRejected
			summary.RejectedRows++
		}
		if err := s.imports.CreateRow(ctx, &record); err != nil {
			return nil, nil, err
		}
		parsedRows = append(parsedRows, record)
		summary.TotalRows++
	}

	duplicateList := make([]string, 0)
	for nick, count := range duplicateEmailOwners {
		if count > 1 {
			duplicateList = append(duplicateList, nick)
		}
	}
	for i := range parsedRows {
		if rowHasDuplicateNick(parsedRows[i], duplicateEmailOwners) {
			parsedRows[i].Status = entity.ImportRowStatusDuplicate
			parsedRows[i].ValidationErrorsJSON = appendValidation(parsedRows[i].ValidationErrorsJSON, "duplicate player email across teams in same import")
			if err := s.imports.UpdateRowStatus(ctx, parsedRows[i].ID, parsedRows[i].Status); err != nil {
				return nil, nil, err
			}
			summary.DuplicateRows++
			if summary.ValidRows > 0 {
				summary.ValidRows--
			}
		}
	}
	summary.DuplicateNicks = duplicateList
	batch.Status = entity.ImportBatchStatusPreviewReady
	batch.SummaryJSON = xjson.MustMarshal(summary)
	if err := s.imports.UpdateBatch(ctx, batch); err != nil {
		return nil, nil, err
	}
	batch, err = s.imports.GetBatch(ctx, batch.ID)
	if err != nil {
		return nil, nil, err
	}
	parsedRows, err = s.imports.ListRowsByBatch(ctx, batch.ID)
	if err != nil {
		return nil, nil, err
	}
	return batch, parsedRows, nil
}

func (s *ImportService) Confirm(ctx context.Context, actorUserID, tournamentID, batchID string) (*entity.ImportBatch, error) {
	ok, err := s.tournaments.CanManageTournament(ctx, tournamentID, actorUserID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, apperror.Forbidden("insufficient tournament permissions")
	}
	batch, err := s.imports.GetBatch(ctx, batchID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("import batch not found")
		}
		return nil, err
	}
	if batch.Status != entity.ImportBatchStatusPreviewReady {
		return nil, apperror.BadRequest("invalid_batch_status", "batch is not ready for confirmation", nil)
	}
	rows, err := s.imports.ListRowsByBatch(ctx, batchID)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		if row.Status != entity.ImportRowStatusValid {
			continue
		}
		createdRowID := row.ID
		approved := false
		team := &entity.Team{ID: uuid.NewString(), TournamentID: tournamentID, Name: row.TeamName, Status: entity.TeamStatusAwaitingConfirmation, ApprovedByManager: &approved, CreatedFromImportRowID: &createdRowID}
		if err := s.teams.CreateTeam(ctx, team); err != nil {
			return nil, err
		}
		tournament, _ := s.tournaments.GetTournament(ctx, tournamentID)
		memberSpecs := []struct {
			Email      string
			Role       string
			IsCaptain  bool
			Substitute bool
		}{
			{Email: row.CaptainNick, Role: entity.MemberRoleCaptain, IsCaptain: true},
			{Email: row.Player2Nick, Role: entity.MemberRolePlayer},
			{Email: row.Player3Nick, Role: entity.MemberRolePlayer},
			{Email: row.Player4Nick, Role: entity.MemberRolePlayer},
			{Email: row.Player5Nick, Role: entity.MemberRolePlayer},
			{Email: row.SubstituteNick, Role: entity.MemberRoleSubstitute, Substitute: true},
		}
		for _, spec := range memberSpecs {
			memberEmail := strings.TrimSpace(spec.Email)
			if memberEmail == "" {
				continue
			}
			nickname := emailPrefix(memberEmail)
			storedEmail := memberEmail
			member := &entity.TeamMember{
				ID:                 uuid.NewString(),
				TeamID:             team.ID,
				Nickname:           nickname,
				Email:              &storedEmail,
				MemberRole:         spec.Role,
				IsCaptain:          spec.IsCaptain,
				IsSubstitute:       spec.Substitute,
				ConfirmationStatus: entity.MemberConfirmationNotFound,
				InvitedAt:          ptrTime(time.Now()),
			}
			matchedUser, err := s.users.GetByEmail(ctx, memberEmail)
			if err == nil {
				member.UserID = &matchedUser.ID
				member.Nickname = matchedUser.Nickname
				if member.Nickname == "" {
					member.Nickname = nickname
				}
				member.ConfirmationStatus = entity.MemberConfirmationPendingConfirmation
			} else if err != pgx.ErrNoRows {
				return nil, err
			}
			if err := s.teams.CreateMember(ctx, member); err != nil {
				return nil, err
			}
			// Send email invite to everyone (even if not registered yet).
			if !spec.IsCaptain && tournament != nil {
				go s.email.SendTeamInvite(memberEmail, team.Name, tournament.Title)
			}
			// Send in-app notification only for registered users.
			if member.UserID != nil && !spec.IsCaptain {
				lang := matchedUser.Lang
				texts := notif.AddedToTeam(lang, team.Name)
				payload := map[string]string{"team_id": team.ID, "team_member_id": member.ID, "team_name": team.Name, "tournament_id": tournamentID}
				if err := s.notifications.Create(ctx, &entity.Notification{
					ID:                uuid.NewString(),
					UserID:            *member.UserID,
					Type:              entity.NotificationAddedToTeam,
					Title:             texts.Title,
					Message:           texts.Message,
					PayloadJSON:       xjson.MustMarshal(payload),
					ActionPayloadJSON: xjson.MustMarshal(payload),
				}); err != nil {
					return nil, err
				}
			}
		}
		if err := s.imports.UpdateRowStatus(ctx, row.ID, entity.ImportRowStatusConfirmed); err != nil {
			return nil, err
		}
	}
	batch.Status = entity.ImportBatchStatusConfirmed
	batch.SummaryJSON = mergeJSONStatus(batch.SummaryJSON, "confirmed")
	if err := s.imports.UpdateBatch(ctx, batch); err != nil {
		return nil, err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &tournamentID, EntityType: "import_batch", EntityID: batchID, ActionType: "import_confirmed", Description: "Import batch confirmed and teams created", MetadataJSON: xjson.MustMarshal(map[string]string{"batch_id": batchID})})
	return s.imports.GetBatch(ctx, batchID)
}

func (s *ImportService) ListBatches(ctx context.Context, actorUserID, tournamentID string) ([]entity.ImportBatch, error) {
	ok, err := s.tournaments.CanManageTournament(ctx, tournamentID, actorUserID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, apperror.Forbidden("insufficient tournament permissions")
	}
	return s.imports.ListBatchesByTournament(ctx, tournamentID)
}

func (s *ImportService) GetBatch(ctx context.Context, actorUserID, batchID string) (*entity.ImportBatch, []entity.ImportRow, error) {
	batch, err := s.imports.GetBatch(ctx, batchID)
	if err != nil {
		return nil, nil, apperror.NotFound("import batch not found")
	}
	ok, err := s.tournaments.CanManageTournament(ctx, batch.TournamentID, actorUserID)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, apperror.Forbidden("insufficient tournament permissions")
	}
	rows, err := s.imports.ListRowsByBatch(ctx, batchID)
	if err != nil {
		return nil, nil, err
	}
	return batch, rows, nil
}

func parseSpreadsheetID(rawURL string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", err
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")

	if len(parts) >= 3 && parts[0] == "spreadsheets" && parts[1] == "d" && parts[2] != "" && parts[2] != "e" {
		return parts[2], nil
	}

	if len(parts) >= 4 && parts[0] == "spreadsheets" && parts[1] == "d" && parts[2] == "e" && parts[3] != "" {
		return parts[3], nil
	}

	return "", fmt.Errorf("spreadsheet id not found in url")
}

func normalizeRow(row []string) []string {
	result := make([]string, 8)
	for i := 0; i < len(result) && i < len(row); i++ {
		result[i] = strings.TrimSpace(row[i])
	}
	return result
}

func toImportRow(batchID string, rowNumber int, cells []string) entity.ImportRow {
	return entity.ImportRow{
		ID:             uuid.NewString(),
		BatchID:        batchID,
		RowNumber:      rowNumber,
		RawDataJSON:    xjson.MustMarshal(cells),
		TeamName:       safeCell(cells, 0),
		Discipline:     safeCell(cells, 1),
		CaptainNick:    safeCell(cells, 2),
		Player2Nick:    safeCell(cells, 3),
		Player3Nick:    safeCell(cells, 4),
		Player4Nick:    safeCell(cells, 5),
		Player5Nick:    safeCell(cells, 6),
		SubstituteNick: safeCell(cells, 7),
	}
}

func isValidEmail(s string) bool {
	at := strings.Index(s, "@")
	return at > 0 && strings.Contains(s[at:], ".")
}

func validateImportRow(row entity.ImportRow) []string {
	errs := make([]string, 0)
	if row.TeamName == "" {
		errs = append(errs, "team_name is required")
	}
	if row.CaptainNick == "" {
		errs = append(errs, "captain_email is required")
	} else if !isValidEmail(row.CaptainNick) {
		errs = append(errs, "captain_email is not a valid email address")
	}
	players := 0
	for _, email := range []string{row.CaptainNick, row.Player2Nick, row.Player3Nick, row.Player4Nick, row.Player5Nick} {
		if e := strings.TrimSpace(email); e != "" {
			if !isValidEmail(e) {
				errs = append(errs, fmt.Sprintf("invalid email: %s", e))
			} else {
				players++
			}
		}
	}
	if players < 4 {
		errs = append(errs, "at least 4 main players including captain are required")
	}
	// validate optional substitute email
	if sub := strings.TrimSpace(row.SubstituteNick); sub != "" && !isValidEmail(sub) {
		errs = append(errs, fmt.Sprintf("invalid substitute email: %s", sub))
	}
	return errs
}

func nonEmptyMembers(row entity.ImportRow) []string {
	members := []string{row.CaptainNick, row.Player2Nick, row.Player3Nick, row.Player4Nick, row.Player5Nick, row.SubstituteNick}
	result := make([]string, 0)
	for _, nick := range members {
		if strings.TrimSpace(nick) != "" {
			result = append(result, nick)
		}
	}
	return result
}

func rowHasDuplicateNick(row entity.ImportRow, dup map[string]int) bool {
	for _, nick := range nonEmptyMembers(row) {
		if dup[strings.ToLower(nick)] > 1 {
			return true
		}
	}
	return false
}

func appendValidation(raw []byte, extra string) []byte {
	current := make([]string, 0)
	_ = json.Unmarshal(raw, &current)
	current = append(current, extra)
	return xjson.MustMarshal(current)
}

func mergeJSONStatus(raw []byte, status string) []byte {
	payload := map[string]interface{}{}
	_ = json.Unmarshal(raw, &payload)
	payload["status"] = status
	return xjson.MustMarshal(payload)
}

func safeCell(cells []string, idx int) string {
	if idx >= 0 && idx < len(cells) {
		return strings.TrimSpace(cells[idx])
	}
	return ""
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
