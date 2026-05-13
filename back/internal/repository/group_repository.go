package repository

import (
	"context"

	"esports-backend/internal/entity"
)

type GroupRepository struct {
	db Queryer
}

func NewGroupRepository(db Queryer) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) CreateGroup(ctx context.Context, g *entity.BracketGroup) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO bracket_groups (id, bracket_id, tournament_id, name, position) VALUES ($1,$2,$3,$4,$5)`,
		g.ID, g.BracketID, g.TournamentID, g.Name, g.Position,
	)
	return err
}

func (r *GroupRepository) CreateMember(ctx context.Context, m *entity.BracketGroupMember) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO bracket_group_members (id, group_id, team_id) VALUES ($1,$2,$3)`,
		m.ID, m.GroupID, m.TeamID,
	)
	return err
}

func (r *GroupRepository) ListByBracketID(ctx context.Context, bracketID string) ([]entity.BracketGroup, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, bracket_id, tournament_id, name, position FROM bracket_groups WHERE bracket_id=$1 ORDER BY position`,
		bracketID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var groups []entity.BracketGroup
	for rows.Next() {
		var g entity.BracketGroup
		if err := rows.Scan(&g.ID, &g.BracketID, &g.TournamentID, &g.Name, &g.Position); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range groups {
		members, err := r.listMembersByGroupID(ctx, groups[i].ID)
		if err != nil {
			return nil, err
		}
		groups[i].Members = members
	}
	return groups, nil
}

func (r *GroupRepository) listMembersByGroupID(ctx context.Context, groupID string) ([]entity.BracketGroupMember, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, group_id, team_id, wins, losses, draws, points, qualified_position
         FROM bracket_group_members WHERE group_id=$1
         ORDER BY points DESC, wins DESC`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []entity.BracketGroupMember
	for rows.Next() {
		var m entity.BracketGroupMember
		if err := rows.Scan(&m.ID, &m.GroupID, &m.TeamID, &m.Wins, &m.Losses, &m.Draws, &m.Points, &m.QualifiedPosition); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *GroupRepository) SetQualifiedPosition(ctx context.Context, groupID, teamID string, position int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE bracket_group_members SET qualified_position=$1 WHERE group_id=$2 AND team_id=$3`,
		position, groupID, teamID,
	)
	return err
}

func (r *GroupRepository) RecordWin(ctx context.Context, groupID, winnerTeamID, loserTeamID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE bracket_group_members SET wins=wins+1, points=points+3 WHERE group_id=$1 AND team_id=$2`,
		groupID, winnerTeamID,
	)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx,
		`UPDATE bracket_group_members SET losses=losses+1 WHERE group_id=$1 AND team_id=$2`,
		groupID, loserTeamID,
	)
	return err
}

func (r *GroupRepository) GetMembersByGroupID(ctx context.Context, groupID string) ([]entity.BracketGroupMember, error) {
	return r.listMembersByGroupID(ctx, groupID)
}
