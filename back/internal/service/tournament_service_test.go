package service

import (
	"context"
	"errors"
	"testing"

	"esports-backend/internal/apperror"
	"esports-backend/internal/entity"
	"esports-backend/internal/repository"

	"github.com/jackc/pgx/v5"
)

// ── stubs ────────────────────────────────────────────────────────────────────

type stubTournamentStore struct {
	tournaments map[string]*entity.Tournament
	roles       map[string][]string // "tournamentID:userID" → roles
	createErr   error
}

func newStubTournamentStore() *stubTournamentStore {
	return &stubTournamentStore{
		tournaments: make(map[string]*entity.Tournament),
		roles:       make(map[string][]string),
	}
}

func (s *stubTournamentStore) Create(_ context.Context, t *entity.Tournament) error {
	if s.createErr != nil {
		return s.createErr
	}
	cp := *t
	s.tournaments[t.ID] = &cp
	return nil
}

func (s *stubTournamentStore) Update(_ context.Context, t *entity.Tournament) error {
	if _, ok := s.tournaments[t.ID]; !ok {
		return pgx.ErrNoRows
	}
	cp := *t
	s.tournaments[t.ID] = &cp
	return nil
}

func (s *stubTournamentStore) SetStatus(_ context.Context, id, status string) error {
	t, ok := s.tournaments[id]
	if !ok {
		return pgx.ErrNoRows
	}
	t.Status = status
	return nil
}

func (s *stubTournamentStore) SoftDelete(_ context.Context, id string) error {
	delete(s.tournaments, id)
	return nil
}

func (s *stubTournamentStore) GetByID(_ context.Context, id string) (*entity.Tournament, error) {
	t, ok := s.tournaments[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *t
	return &cp, nil
}

func (s *stubTournamentStore) GetBySlug(_ context.Context, _ string) (*entity.Tournament, error) {
	return nil, pgx.ErrNoRows
}

func (s *stubTournamentStore) ListPublic(_ context.Context, _, _ int, _ repository.TournamentFilter) ([]entity.Tournament, error) {
	out := make([]entity.Tournament, 0, len(s.tournaments))
	for _, t := range s.tournaments {
		if t.Visibility == entity.TournamentVisibilityPublic {
			out = append(out, *t)
		}
	}
	return out, nil
}

func (s *stubTournamentStore) CountPublic(_ context.Context, _ repository.TournamentFilter) (int, error) {
	n := 0
	for _, t := range s.tournaments {
		if t.Visibility == entity.TournamentVisibilityPublic {
			n++
		}
	}
	return n, nil
}

func (s *stubTournamentStore) ListAll(_ context.Context, _, _ int) ([]entity.Tournament, error) {
	out := make([]entity.Tournament, 0, len(s.tournaments))
	for _, t := range s.tournaments {
		out = append(out, *t)
	}
	return out, nil
}

func (s *stubTournamentStore) CountAll(_ context.Context) (int, error) {
	return len(s.tournaments), nil
}

func (s *stubTournamentStore) AddRole(_ context.Context, r *entity.TournamentUserRole) error {
	key := r.TournamentID + ":" + r.UserID
	s.roles[key] = append(s.roles[key], r.Role)
	return nil
}

func (s *stubTournamentStore) RemoveRole(_ context.Context, tournamentID, userID, role string) error {
	key := tournamentID + ":" + userID
	existing := s.roles[key]
	filtered := existing[:0]
	for _, r := range existing {
		if r != role {
			filtered = append(filtered, r)
		}
	}
	s.roles[key] = filtered
	return nil
}

func (s *stubTournamentStore) HasRole(_ context.Context, tournamentID, userID, role string) (bool, error) {
	key := tournamentID + ":" + userID
	for _, r := range s.roles[key] {
		if r == role {
			return true, nil
		}
	}
	return false, nil
}

func (s *stubTournamentStore) ListRoles(_ context.Context, tournamentID, userID string) ([]string, error) {
	return s.roles[tournamentID+":"+userID], nil
}

func (s *stubTournamentStore) ListUserIDsByRoles(_ context.Context, _ string, _ []string) ([]string, error) {
	return nil, nil
}

type stubTeamStore struct{}

func (s *stubTeamStore) ListByTournament(_ context.Context, _ string, _ bool) ([]entity.Team, error) {
	return nil, nil
}

type stubBracketStore struct{}

func (s *stubBracketStore) ListMatchesByTournament(_ context.Context, _ string) ([]entity.Match, error) {
	return nil, nil
}

type stubAuditStore struct{}

func (s *stubAuditStore) Create(_ context.Context, _ *entity.AuditLog) error { return nil }

// ── helpers ───────────────────────────────────────────────────────────────────

func newSvc() (*TournamentService, *stubTournamentStore) {
	store := newStubTournamentStore()
	svc := NewTournamentService(store, &stubTeamStore{}, &stubBracketStore{}, &stubAuditStore{})
	return svc, store
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestCreate_SetsOwnerRoleAndDraftStatus(t *testing.T) {
	svc, store := newSvc()
	ctx := context.Background()

	tournament, err := svc.Create(ctx, "user-1", CreateTournamentInput{
		Title:      "Test Cup",
		Visibility: entity.TournamentVisibilityPublic,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tournament.Status != entity.TournamentStatusDraft {
		t.Errorf("expected status %q, got %q", entity.TournamentStatusDraft, tournament.Status)
	}
	if tournament.OwnerUserID != "user-1" {
		t.Errorf("expected owner user-1, got %q", tournament.OwnerUserID)
	}

	hasOwner, _ := store.HasRole(ctx, tournament.ID, "user-1", entity.TournamentRoleOwner)
	if !hasOwner {
		t.Error("expected owner role to be assigned")
	}
}

func TestCreate_DefaultsRegistrationModeToTeam(t *testing.T) {
	svc, _ := newSvc()
	tournament, err := svc.Create(context.Background(), "user-1", CreateTournamentInput{
		Title:      "Test Cup",
		Visibility: entity.TournamentVisibilityPublic,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tournament.RegistrationMode != "team" {
		t.Errorf("expected default registration_mode=team, got %q", tournament.RegistrationMode)
	}
}

func TestCreate_ReturnsErrorOnStoreFailure(t *testing.T) {
	svc, store := newSvc()
	store.createErr = errors.New("db error")

	_, err := svc.Create(context.Background(), "user-1", CreateTournamentInput{Title: "X", Visibility: "public"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPublic_ReturnsPublicTournament(t *testing.T) {
	svc, store := newSvc()
	ctx := context.Background()
	store.tournaments["t1"] = &entity.Tournament{
		ID:         "t1",
		Visibility: entity.TournamentVisibilityPublic,
	}

	got, err := svc.GetPublic(ctx, "t1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "t1" {
		t.Errorf("expected tournament t1, got %q", got.ID)
	}
}

func TestGetPublic_PrivateTournament_DeniesAnonymous(t *testing.T) {
	svc, store := newSvc()
	store.tournaments["t1"] = &entity.Tournament{
		ID:         "t1",
		Visibility: entity.TournamentVisibilityPrivate,
	}

	_, err := svc.GetPublic(context.Background(), "t1", "")
	if err == nil {
		t.Fatal("expected forbidden error, got nil")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.HTTPStatus != 403 {
		t.Errorf("expected 403 AppError, got %v", err)
	}
}

func TestGetPublic_PrivateTournament_AllowsMember(t *testing.T) {
	svc, store := newSvc()
	ctx := context.Background()
	store.tournaments["t1"] = &entity.Tournament{
		ID:         "t1",
		Visibility: entity.TournamentVisibilityPrivate,
	}
	store.roles["t1:user-1"] = []string{entity.TournamentRoleManager}

	got, err := svc.GetPublic(ctx, "t1", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "t1" {
		t.Errorf("expected tournament t1, got %q", got.ID)
	}
}

func TestGetPublic_NotFound(t *testing.T) {
	svc, _ := newSvc()
	_, err := svc.GetPublic(context.Background(), "missing", "")
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.HTTPStatus != 404 {
		t.Errorf("expected 404 AppError, got %v", err)
	}
}

func TestCanManageTournament_OwnerCanManage(t *testing.T) {
	svc, store := newSvc()
	store.roles["t1:user-1"] = []string{entity.TournamentRoleOwner}

	ok, err := svc.CanManageTournament(context.Background(), "t1", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected owner to be able to manage")
	}
}

func TestCanManageTournament_ManagerCanManage(t *testing.T) {
	svc, store := newSvc()
	store.roles["t1:user-2"] = []string{entity.TournamentRoleManager}

	ok, _ := svc.CanManageTournament(context.Background(), "t1", "user-2")
	if !ok {
		t.Error("expected manager to be able to manage")
	}
}

func TestCanManageTournament_StrangerCannotManage(t *testing.T) {
	svc, _ := newSvc()

	ok, _ := svc.CanManageTournament(context.Background(), "t1", "stranger")
	if ok {
		t.Error("expected stranger to be denied management")
	}
}

func TestAddManager_OwnerCanAdd(t *testing.T) {
	svc, store := newSvc()
	ctx := context.Background()
	store.roles["t1:owner"] = []string{entity.TournamentRoleOwner}

	if err := svc.AddManager(ctx, "owner", "t1", "new-manager"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hasRole, _ := store.HasRole(ctx, "t1", "new-manager", entity.TournamentRoleManager)
	if !hasRole {
		t.Error("expected manager role to be assigned")
	}
}

func TestAddManager_NonOwnerForbidden(t *testing.T) {
	svc, store := newSvc()
	store.roles["t1:manager"] = []string{entity.TournamentRoleManager}

	err := svc.AddManager(context.Background(), "manager", "t1", "someone")
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.HTTPStatus != 403 {
		t.Errorf("expected 403, got %v", err)
	}
}

func TestDelete_OwnerCanDelete(t *testing.T) {
	svc, store := newSvc()
	ctx := context.Background()
	store.tournaments["t1"] = &entity.Tournament{ID: "t1"}
	store.roles["t1:owner"] = []string{entity.TournamentRoleOwner}

	if err := svc.Delete(ctx, "owner", "t1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, exists := store.tournaments["t1"]; exists {
		t.Error("expected tournament to be deleted")
	}
}

func TestDelete_StrangerForbidden(t *testing.T) {
	svc, store := newSvc()
	store.tournaments["t1"] = &entity.Tournament{ID: "t1"}

	err := svc.Delete(context.Background(), "stranger", "t1")
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.HTTPStatus != 403 {
		t.Errorf("expected 403, got %v", err)
	}
}

func TestListPublic_ReturnsOnlyPublicTournaments(t *testing.T) {
	svc, store := newSvc()
	store.tournaments["pub"] = &entity.Tournament{ID: "pub", Visibility: entity.TournamentVisibilityPublic}
	store.tournaments["priv"] = &entity.Tournament{ID: "priv", Visibility: entity.TournamentVisibilityPrivate}

	items, total, err := svc.ListPublic(context.Background(), 20, 0, repository.TournamentFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ID != "pub" {
		t.Errorf("expected 1 public tournament, got %d", len(items))
	}
	if total != 1 {
		t.Errorf("expected total=1, got %d", total)
	}
}
