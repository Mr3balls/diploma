package bracket

import (
	"fmt"
	"sort"
)

// MatchOutcome stores the result of one completed match for standings computation.
type MatchOutcome struct {
	NodeIndex  int // matches a Node.Index in the bracket
	WinnerSeed int // 1-based winning seed; 0 means BYE auto-advance
	LoserSeed  int // 1-based losing seed; 0 means BYE (no real opponent)
}

// Standing represents one participant's final tournament placement.
type Standing struct {
	Rank   int    // 1-based placement; shared rank means tied
	Tied   bool   // true when tied with the immediately following entry
	Seed   int    // 1-based seed
	Wins   int
	Losses int
	// WinsLosses returns a "W-L" formatted string for display.
}

func (s Standing) WinsLosses() string {
	return fmt.Sprintf("%d-%d", s.Wins, s.Losses)
}

// ComputeStandings returns standings sorted from 1st place to last.
//
// n is the number of real participants (excluding BYEs).
// format is "single_elimination" or "double_elimination".
//
// For single elimination: placement determined by which WB round the participant lost.
// For double elimination: placement determined by which LB/GF round they were eliminated.
// BYE slots (seed 0) are not included in the output.
func ComputeStandings(nodes []Node, outcomes []MatchOutcome, n int, format string) []Standing {
	byIdx := make(map[int]*Node, len(nodes))
	for i := range nodes {
		byIdx[nodes[i].Index] = &nodes[i]
	}

	wins := make(map[int]int)
	losses := make(map[int]int)
	// eliminatedIn: seed → the match node where the participant was FINALLY eliminated.
	// For SE: any loss. For DE: loss in LB or GF only (WB loss → drops to LB).
	eliminatedIn := make(map[int]*Node)

	for _, o := range outcomes {
		nd := byIdx[o.NodeIndex]
		if nd == nil {
			continue
		}
		if o.WinnerSeed > 0 {
			wins[o.WinnerSeed]++
		}
		if o.LoserSeed > 0 {
			losses[o.LoserSeed]++
			switch format {
			case "single_elimination":
				eliminatedIn[o.LoserSeed] = nd
			default: // double_elimination
				// WB loser drops to LB — not eliminated yet.
				if nd.Section != SectionWB {
					eliminatedIn[o.LoserSeed] = nd
				}
			}
		}
	}

	// sortKey: higher = better placement.
	// SE:  WB round number (Final = roundsWB, champion not eliminated = roundsWB+1).
	// DE:  LB round number for LB eliminations; 9000 for GF loser; 10000 for champion.
	sortKey := func(seed int) int {
		nd := eliminatedIn[seed]
		if nd == nil {
			return 10000 // champion: no elimination match recorded
		}
		switch format {
		case "single_elimination":
			return nd.Round
		default:
			if nd.Section == SectionGF {
				return 9000 // always rank 2
			}
			return nd.Round // LB round; higher = survived longer
		}
	}

	type entry struct{ seed, key int }
	entries := make([]entry, 0, n)
	for seed := 1; seed <= n; seed++ {
		entries = append(entries, entry{seed, sortKey(seed)})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].key != entries[j].key {
			return entries[i].key > entries[j].key
		}
		return entries[i].seed < entries[j].seed
	})

	result := make([]Standing, len(entries))
	for i, e := range entries {
		rank := i + 1
		if i > 0 && entries[i].key == entries[i-1].key {
			rank = result[i-1].Rank // same rank as tied predecessor
		}
		result[i] = Standing{
			Rank:   rank,
			Seed:   e.seed,
			Wins:   wins[e.seed],
			Losses: losses[e.seed],
		}
	}
	// Mark Tied flag for consecutive same-rank entries.
	for i := 0; i < len(result)-1; i++ {
		if result[i].Rank == result[i+1].Rank {
			result[i].Tied = true
		}
	}

	return result
}
