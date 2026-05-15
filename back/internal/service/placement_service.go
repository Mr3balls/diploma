package service

import (
	"context"
	"sort"

	"esports-backend/internal/apperror"
	"esports-backend/internal/entity"
)

type TeamPlacement struct {
	TeamID    string `json:"team_id"`
	TeamName  string `json:"team_name"`
	PlaceFrom int    `json:"place_from"`
	PlaceTo   int    `json:"place_to"`
	IsActive  bool   `json:"is_active"`
}

func (s *BracketService) ComputePlacements(ctx context.Context, tournamentID string) ([]TeamPlacement, error) {
	bracket, err := s.brackets.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, apperror.NotFound("bracket not found")
	}

	matches, err := s.brackets.ListMatchesByTournament(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	// Build a unified name map: covers both team-based and individual tournaments.
	names := make(map[string]string)

	teams, _ := s.teams.ListByTournament(ctx, tournamentID, false)
	for _, t := range teams {
		names[t.ID] = t.Name
	}

	participants, _ := s.participants.ListByTournament(ctx, tournamentID)
	for _, p := range participants {
		names[p.ID] = p.Name
	}

	switch bracket.Format {
	case "double_elimination":
		return computeDEPlacements(matches, names), nil
	case "group_stage", "group_de":
		return computeGroupPlacements(matches, names), nil
	default:
		return computeSEPlacements(matches, names), nil
	}
}

// ── match field helpers (team-based or participant-based) ─────────────────────

func matchPlayer1(m entity.Match) *string {
	if m.Team1ID != nil {
		return m.Team1ID
	}
	return m.Participant1ID
}

func matchPlayer2(m entity.Match) *string {
	if m.Team2ID != nil {
		return m.Team2ID
	}
	return m.Participant2ID
}

func matchWinner(m entity.Match) *string {
	if m.WinnerTeamID != nil {
		return m.WinnerTeamID
	}
	return m.WinnerParticipantID
}

func matchLoser(m entity.Match) *string {
	winner := matchWinner(m)
	if winner == nil {
		return nil
	}
	if p1 := matchPlayer1(m); p1 != nil && *p1 != *winner {
		return p1
	}
	if p2 := matchPlayer2(m); p2 != nil && *p2 != *winner {
		return p2
	}
	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

// buildPlacements turns a score→[]teamID map into an ordered []TeamPlacement.
// Scores are sorted descending (higher = better placement).
// Negative scores go last and are marked IsActive=true.
func buildPlacements(scoreMap map[int][]string, names map[string]string) []TeamPlacement {
	var positiveScores []int
	var negativeScores []int
	for s := range scoreMap {
		if s >= 0 {
			positiveScores = append(positiveScores, s)
		} else {
			negativeScores = append(negativeScores, s)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(positiveScores)))
	sort.Sort(sort.Reverse(sort.IntSlice(negativeScores)))

	var result []TeamPlacement
	place := 1

	for _, s := range positiveScores {
		group := scoreMap[s]
		sort.Strings(group)
		n := len(group)
		for _, id := range group {
			result = append(result, TeamPlacement{
				TeamID: id, TeamName: names[id],
				PlaceFrom: place, PlaceTo: place + n - 1,
				IsActive: false,
			})
		}
		place += n
	}
	for _, s := range negativeScores {
		group := scoreMap[s]
		sort.Strings(group)
		n := len(group)
		for _, id := range group {
			result = append(result, TeamPlacement{
				TeamID: id, TeamName: names[id],
				PlaceFrom: place, PlaceTo: place + n - 1,
				IsActive: true,
			})
		}
		place += n
	}
	return result
}

// ── Single Elimination ────────────────────────────────────────────────────────

func computeSEPlacements(matches []entity.Match, names map[string]string) []TeamPlacement {
	totalRounds := 0
	appeared := make(map[string]bool)

	for _, m := range matches {
		if m.IsBye || m.GroupID != nil {
			continue
		}
		if p1 := matchPlayer1(m); p1 != nil {
			appeared[*p1] = true
		}
		if p2 := matchPlayer2(m); p2 != nil {
			appeared[*p2] = true
		}
		if m.RoundNumber > totalRounds {
			totalRounds = m.RoundNumber
		}
	}
	if totalRounds == 0 {
		return nil
	}

	scoreMap := make(map[int][]string)
	eliminated := make(map[string]int)

	for _, m := range matches {
		w := matchWinner(m)
		if m.IsBye || m.GroupID != nil || m.Status != "finished" || w == nil {
			continue
		}
		r := m.RoundNumber
		if r == totalRounds {
			eliminated[*w] = totalRounds + 1
		}
		if id := matchLoser(m); id != nil {
			eliminated[*id] = r
		}
	}

	for id := range appeared {
		if s, ok := eliminated[id]; ok {
			scoreMap[s] = append(scoreMap[s], id)
		} else {
			scoreMap[-1] = append(scoreMap[-1], id)
		}
	}

	return buildPlacements(scoreMap, names)
}

// ── Double Elimination ────────────────────────────────────────────────────────

func computeDEPlacements(matches []entity.Match, names map[string]string) []TeamPlacement {
	appeared := make(map[string]bool)

	gfMaxRound := 0
	for _, m := range matches {
		if !m.IsBye && m.BracketSection == "GF" && m.RoundNumber > gfMaxRound {
			gfMaxRound = m.RoundNumber
		}
	}

	lbMaxRound := 0
	for _, m := range matches {
		if !m.IsBye && m.BracketSection == "LB" && m.RoundNumber > lbMaxRound {
			lbMaxRound = m.RoundNumber
		}
	}

	eliminated := make(map[string]int)

	for _, m := range matches {
		if m.IsBye {
			continue
		}
		if p1 := matchPlayer1(m); p1 != nil {
			appeared[*p1] = true
		}
		if p2 := matchPlayer2(m); p2 != nil {
			appeared[*p2] = true
		}
		w := matchWinner(m)
		if m.Status != "finished" || w == nil {
			continue
		}

		switch m.BracketSection {
		case "GF":
			if m.RoundNumber == gfMaxRound {
				eliminated[*w] = lbMaxRound + 2
				if id := matchLoser(m); id != nil {
					eliminated[*id] = lbMaxRound + 1
				}
			}
		case "LB":
			if id := matchLoser(m); id != nil {
				if _, already := eliminated[*id]; !already {
					eliminated[*id] = m.RoundNumber
				}
			}
		}
	}

	scoreMap := make(map[int][]string)
	for id := range appeared {
		if s, ok := eliminated[id]; ok {
			scoreMap[s] = append(scoreMap[s], id)
		} else {
			scoreMap[-1] = append(scoreMap[-1], id)
		}
	}

	return buildPlacements(scoreMap, names)
}

// ── Group Stage / Group DE ────────────────────────────────────────────────────
// Strategy: playoff matches (no group_id) use SE logic for top spots.
// Non-playoff teams are ranked by how many group matches they won.

func computeGroupPlacements(matches []entity.Match, names map[string]string) []TeamPlacement {
	var playoffMatches []entity.Match
	groupWins := make(map[string]int)
	groupAppeared := make(map[string]bool)

	for _, m := range matches {
		if m.IsBye {
			continue
		}
		if m.GroupID == nil {
			playoffMatches = append(playoffMatches, m)
		} else {
			if p1 := matchPlayer1(m); p1 != nil {
				groupAppeared[*p1] = true
			}
			if p2 := matchPlayer2(m); p2 != nil {
				groupAppeared[*p2] = true
			}
			if w := matchWinner(m); m.Status == "finished" && w != nil {
				groupWins[*w]++
			}
		}
	}

	// Teams in playoff
	playoffTeams := make(map[string]bool)
	for _, m := range playoffMatches {
		if p1 := matchPlayer1(m); p1 != nil {
			playoffTeams[*p1] = true
		}
		if p2 := matchPlayer2(m); p2 != nil {
			playoffTeams[*p2] = true
		}
	}

	// Playoff placements via SE logic
	playoffPlacements := computeSEPlacements(playoffMatches, names)

	// Rank non-playoff teams by group wins
	var nonPlayoff []string
	for id := range groupAppeared {
		if !playoffTeams[id] {
			nonPlayoff = append(nonPlayoff, id)
		}
	}
	sort.Slice(nonPlayoff, func(i, j int) bool {
		if groupWins[nonPlayoff[i]] != groupWins[nonPlayoff[j]] {
			return groupWins[nonPlayoff[i]] > groupWins[nonPlayoff[j]]
		}
		return nonPlayoff[i] < nonPlayoff[j]
	})

	// Build result: playoff first, then non-playoff
	result := make([]TeamPlacement, 0, len(playoffPlacements)+len(nonPlayoff))
	result = append(result, playoffPlacements...)

	startPlace := 1
	if len(result) > 0 {
		startPlace = result[len(result)-1].PlaceTo + 1
	}

	// Group non-playoff teams by same wins count for shared places
	i := 0
	for i < len(nonPlayoff) {
		wins := groupWins[nonPlayoff[i]]
		j := i
		for j < len(nonPlayoff) && groupWins[nonPlayoff[j]] == wins {
			j++
		}
		n := j - i
		for _, id := range nonPlayoff[i:j] {
			result = append(result, TeamPlacement{
				TeamID: id, TeamName: names[id],
				PlaceFrom: startPlace, PlaceTo: startPlace + n - 1,
				IsActive: false,
			})
		}
		startPlace += n
		i = j
	}

	return result
}
