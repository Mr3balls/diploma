package service

import (
	"context"
	"time"

	"esports-backend/internal/apperror"
	"esports-backend/internal/entity"
	"esports-backend/internal/pkg/xjson"
	"esports-backend/internal/repository"

	"github.com/google/uuid"
)

type TournamentService struct {
	tournaments repository.TournamentStore
	teams       repository.TeamStore
	brackets    repository.BracketStore
	audits      repository.AuditStore
}

func NewTournamentService(tournaments repository.TournamentStore, teams repository.TeamStore, brackets repository.BracketStore, audits repository.AuditStore) *TournamentService {
	return &TournamentService{tournaments: tournaments, teams: teams, brackets: brackets, audits: audits}
}

type CreateTournamentInput struct {
	Title                string
	Discipline           string
	Description          *string
	Rules                *string
	Location             *string
	MaxTeams             int
	Format               string
	GroupCount           *int
	RegistrationDeadline *time.Time
	StartAt              *time.Time
	Visibility           string
	RegistrationMode     string // "team" | "individual"
}

func (s *TournamentService) ListPublic(ctx context.Context, limit, offset int, f repository.TournamentFilter) ([]entity.Tournament, int, error) {
	items, err := s.tournaments.ListPublic(ctx, limit, offset, f)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.tournaments.CountPublic(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *TournamentService) GetPublic(ctx context.Context, tournamentID, requesterID string) (*entity.Tournament, error) {
	t, err := s.tournaments.GetByID(ctx, tournamentID)
	if err != nil {
		return nil, apperror.NotFound("tournament not found")
	}
	if t.Visibility == entity.TournamentVisibilityPrivate {
		if requesterID == "" {
			return nil, apperror.Forbidden("private tournament")
		}
		allowed, err := s.CanReadTournament(ctx, tournamentID, requesterID)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, apperror.Forbidden("private tournament")
		}
	}
	return t, nil
}

func (s *TournamentService) Create(ctx context.Context, ownerUserID string, in CreateTournamentInput) (*entity.Tournament, error) {
	mode := in.RegistrationMode
	if mode == "" {
		mode = "team"
	}
	tournament := &entity.Tournament{
		ID:                   uuid.NewString(),
		Title:                in.Title,
		Discipline:           in.Discipline,
		Description:          in.Description,
		Rules:                in.Rules,
		Location:             in.Location,
		MaxTeams:             in.MaxTeams,
		Format:               in.Format,
		GroupCount:           in.GroupCount,
		Status:               entity.TournamentStatusDraft,
		Visibility:           in.Visibility,
		OwnerUserID:          ownerUserID,
		RegistrationDeadline: in.RegistrationDeadline,
		StartAt:              in.StartAt,
		RegistrationMode:     mode,
	}

	if err := s.tournaments.Create(ctx, tournament); err != nil {
		return nil, err
	}
	if err := s.tournaments.AddRole(ctx, &entity.TournamentUserRole{ID: uuid.NewString(), TournamentID: tournament.ID, UserID: ownerUserID, Role: entity.TournamentRoleOwner, AssignedBy: ownerUserID}); err != nil {
		return nil, err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &ownerUserID, TournamentID: &tournament.ID, EntityType: "tournament", EntityID: tournament.ID, ActionType: "tournament_created", Description: "Tournament created", MetadataJSON: xjson.MustMarshal(tournament)})
	return s.tournaments.GetByID(ctx, tournament.ID)
}

func (s *TournamentService) Update(ctx context.Context, actorUserID string, tournament *entity.Tournament) (*entity.Tournament, error) {
	ok, err := s.CanManageTournament(ctx, tournament.ID, actorUserID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, apperror.Forbidden("insufficient tournament permissions")
	}
	if err := s.tournaments.Update(ctx, tournament); err != nil {
		return nil, err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &tournament.ID, EntityType: "tournament", EntityID: tournament.ID, ActionType: "tournament_updated", Description: "Tournament updated", MetadataJSON: xjson.MustMarshal(tournament)})
	return s.tournaments.GetByID(ctx, tournament.ID)
}

func (s *TournamentService) Delete(ctx context.Context, actorUserID, tournamentID string) error {
	ok, err := s.CanManageTournament(ctx, tournamentID, actorUserID)
	if err != nil {
		return err
	}
	if !ok {
		return apperror.Forbidden("insufficient tournament permissions")
	}
	if err := s.tournaments.SoftDelete(ctx, tournamentID); err != nil {
		return err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &tournamentID, EntityType: "tournament", EntityID: tournamentID, ActionType: "tournament_deleted", Description: "Tournament deleted", MetadataJSON: xjson.MustMarshal(map[string]string{"tournament_id": tournamentID})})
	return nil
}

func (s *TournamentService) ChangeStatus(ctx context.Context, actorUserID, tournamentID, status string) error {
	ok, err := s.CanManageTournament(ctx, tournamentID, actorUserID)
	if err != nil {
		return err
	}
	if !ok {
		return apperror.Forbidden("insufficient tournament permissions")
	}
	if err := s.tournaments.SetStatus(ctx, tournamentID, status); err != nil {
		return err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &tournamentID, EntityType: "tournament", EntityID: tournamentID, ActionType: "status_changed", Description: "Tournament status changed", MetadataJSON: xjson.MustMarshal(map[string]string{"status": status})})
	return nil
}

func (s *TournamentService) AddManager(ctx context.Context, actorUserID, tournamentID, managerUserID string) error {
	isOwner, err := s.tournaments.HasRole(ctx, tournamentID, actorUserID, entity.TournamentRoleOwner)
	if err != nil {
		return err
	}
	if !isOwner {
		return apperror.Forbidden("only owner can assign managers")
	}
	err = s.tournaments.AddRole(ctx, &entity.TournamentUserRole{ID: uuid.NewString(), TournamentID: tournamentID, UserID: managerUserID, Role: entity.TournamentRoleManager, AssignedBy: actorUserID})
	if err != nil {
		return err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &tournamentID, EntityType: "tournament_user_role", EntityID: managerUserID, ActionType: "manager_added", Description: "Manager added to tournament", MetadataJSON: xjson.MustMarshal(map[string]string{"manager_user_id": managerUserID})})
	return nil
}

func (s *TournamentService) RemoveManager(ctx context.Context, actorUserID, tournamentID, managerUserID string) error {
	isOwner, err := s.tournaments.HasRole(ctx, tournamentID, actorUserID, entity.TournamentRoleOwner)
	if err != nil {
		return err
	}
	if !isOwner {
		return apperror.Forbidden("only owner can remove managers")
	}
	if err := s.tournaments.RemoveRole(ctx, tournamentID, managerUserID, entity.TournamentRoleManager); err != nil {
		return err
	}
	_ = s.audits.Create(ctx, &entity.AuditLog{ID: uuid.NewString(), ActorUserID: &actorUserID, TournamentID: &tournamentID, EntityType: "tournament_user_role", EntityID: managerUserID, ActionType: "manager_removed", Description: "Manager removed from tournament", MetadataJSON: xjson.MustMarshal(map[string]string{"manager_user_id": managerUserID})})
	return nil
}

func (s *TournamentService) CanManageTournament(ctx context.Context, tournamentID, userID string) (bool, error) {
	roles, err := s.tournaments.ListRoles(ctx, tournamentID, userID)
	if err != nil {
		return false, err
	}
	for _, role := range roles {
		if role == entity.TournamentRoleOwner || role == entity.TournamentRoleManager {
			return true, nil
		}
	}
	return false, nil
}

func (s *TournamentService) CanReadTournament(ctx context.Context, tournamentID, userID string) (bool, error) {
	tournament, err := s.tournaments.GetByID(ctx, tournamentID)
	if err != nil {
		return false, err
	}
	if tournament.Visibility == entity.TournamentVisibilityPublic {
		return true, nil
	}
	roles, err := s.tournaments.ListRoles(ctx, tournamentID, userID)
	if err != nil {
		return false, err
	}
	return len(roles) > 0, nil
}

func (s *TournamentService) GetTournament(ctx context.Context, tournamentID string) (*entity.Tournament, error) {
	t, err := s.tournaments.GetByID(ctx, tournamentID)
	if err != nil {
		return nil, apperror.NotFound("tournament not found")
	}
	return t, nil
}

func (s *TournamentService) ListTournamentTeams(ctx context.Context, tournamentID string, admin bool) ([]entity.Team, error) {
	return s.teams.ListByTournament(ctx, tournamentID, admin)
}

func (s *TournamentService) ListTournamentMatches(ctx context.Context, tournamentID string) ([]entity.Match, error) {
	return s.brackets.ListMatchesByTournament(ctx, tournamentID)
}
