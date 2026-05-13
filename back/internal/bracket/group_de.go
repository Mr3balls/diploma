package bracket

import "fmt"

// BuildGroupDE constructs a group-phase double-elimination bracket without GF.
// Used within each group of the "group_de" tournament format.
//
// For a group of n participants (bracket size S = nextPow2(n)):
//
//	WB: all rounds through WB Final (S-1 matches).
//	LB: rounds R1 through R(2*(roundsWB-1)-1) — all except the last major round
//	    that would pair the WB Final loser vs the last LB minor winner.
//
// Three seeds are extracted from each group:
//
//	Seed 1 — WB Final winner  → advances to playoff Semifinal (bye in QF)
//	Seed 2 — WB Final loser   → advances to playoff Quarterfinal
//	Seed 3 — LB Final winner  → advances to playoff Quarterfinal
//
// The WB Final's LoseNext is -1: the loser is extracted as Seed 2 directly
// without entering the LB. lbFinalIdx is the index of the last LB minor round
// match (LB Final for this group). Returns -1 for lbFinalIdx when n < 4.
func BuildGroupDE(n int) (nodes []Node, wbFinalIdx int, lbFinalIdx int, err error) {
	if n < 4 {
		return nil, 0, 0, fmt.Errorf("bracket: group_de requires at least 4 participants per group, got %d", n)
	}

	size := nextPow2(n)
	roundsWB := intLog2(size)
	seeds := StandardSeeds(size)

	nodes = make([]Node, 0, 2*size)
	idx := 0

	byRoundWB := make([][]int, roundsWB+1)

	// ── WB Round 1 ────────────────────────────────────────────────────────
	r1count := size / 2
	byRoundWB[1] = make([]int, r1count)
	for i := 0; i < r1count; i++ {
		s1, s2 := seeds[2*i], seeds[2*i+1]
		if s1 > n {
			s1 = 0
		}
		if s2 > n {
			s2 = 0
		}
		nodes = append(nodes, Node{
			Index:    idx,
			Section:  SectionWB,
			Round:    1,
			Slot:     i + 1,
			Seed1:    s1,
			Seed2:    s2,
			Src1:     -1,
			Src2:     -1,
			WinNext:  -1,
			LoseNext: -1,
			IsBye:    s1 == 0 || s2 == 0,
		})
		byRoundWB[1][i] = idx
		idx++
	}

	// ── WB Rounds 2…roundsWB ──────────────────────────────────────────────
	for r := 2; r <= roundsWB; r++ {
		prev := byRoundWB[r-1]
		count := len(prev) / 2
		byRoundWB[r] = make([]int, count)
		for i := 0; i < count; i++ {
			left, right := prev[2*i], prev[2*i+1]
			nodes = append(nodes, Node{
				Index:    idx,
				Section:  SectionWB,
				Round:    r,
				Slot:     i + 1,
				Src1:     left,
				Src2:     right,
				WinNext:  -1,
				LoseNext: -1,
			})
			nodes[left].WinNext = idx
			nodes[left].WinSlot = 1
			nodes[right].WinNext = idx
			nodes[right].WinSlot = 2
			byRoundWB[r][i] = idx
			idx++
		}
	}
	wbFinalIdx = byRoundWB[roundsWB][0]
	// WB Final loser is extracted as Seed 2 — does NOT enter LB (LoseNext = -1).

	// ── LB: all rounds except last major ─────────────────────────────────
	// Standard lbRounds = 2*(roundsWB-1).
	// In group_de we stop one round earlier (exclude the last LB major that
	// would receive the WB Final loser). The resulting last round is always a
	// minor round — its winner is group Seed 3.
	lbRoundsTotal := 2 * (roundsWB - 1)
	lbRoundsGroupDE := lbRoundsTotal - 1 // always odd (last minor round)

	byRoundLB := make([][]int, lbRoundsGroupDE+1)
	lbFinalIdx = -1

	if lbRoundsGroupDE >= 1 {
		// LBR1 (minor): fold WBR1 losers.
		lbR1count := r1count / 2
		byRoundLB[1] = make([]int, lbR1count)
		wbR1 := byRoundWB[1]
		for i := 0; i < lbR1count; i++ {
			wbLeft := wbR1[i]
			wbRight := wbR1[r1count-1-i]
			nodes = append(nodes, Node{
				Index:    idx,
				Section:  SectionLB,
				Round:    1,
				Slot:     i + 1,
				Src1:     -1,
				Src2:     -1,
				WinNext:  -1,
				LoseNext: -1,
			})
			nodes[wbLeft].LoseNext = idx
			nodes[wbLeft].LoseSlot = 1
			nodes[wbRight].LoseNext = idx
			nodes[wbRight].LoseSlot = 2
			byRoundLB[1][i] = idx
			idx++
		}

		// LBR2…lbRoundsGroupDE
		for lbr := 2; lbr <= lbRoundsGroupDE; lbr++ {
			isMajor := lbr%2 == 0
			if isMajor {
				k := lbr / 2
				wbr := k + 1
				wbLosers := byRoundWB[wbr]
				lbPrev := byRoundLB[lbr-1]
				count := len(lbPrev)
				byRoundLB[lbr] = make([]int, count)
				for i := 0; i < count; i++ {
					wbLoserIdx := wbLosers[i]
					if k%2 == 1 && count > 1 {
						wbLoserIdx = wbLosers[count-1-i]
					}
					lbWinnerIdx := lbPrev[i]
					nodes = append(nodes, Node{
						Index:    idx,
						Section:  SectionLB,
						Round:    lbr,
						Slot:     i + 1,
						Src1:     lbWinnerIdx,
						Src2:     -1,
						WinNext:  -1,
						LoseNext: -1,
					})
					nodes[lbWinnerIdx].WinNext = idx
					nodes[lbWinnerIdx].WinSlot = 1
					nodes[wbLoserIdx].LoseNext = idx
					nodes[wbLoserIdx].LoseSlot = 2
					byRoundLB[lbr][i] = idx
					idx++
				}
			} else {
				lbPrev := byRoundLB[lbr-1]
				count := len(lbPrev) / 2
				byRoundLB[lbr] = make([]int, count)
				for i := 0; i < count; i++ {
					left, right := lbPrev[2*i], lbPrev[2*i+1]
					nodes = append(nodes, Node{
						Index:    idx,
						Section:  SectionLB,
						Round:    lbr,
						Slot:     i + 1,
						Src1:     left,
						Src2:     right,
						WinNext:  -1,
						LoseNext: -1,
					})
					nodes[left].WinNext = idx
					nodes[left].WinSlot = 1
					nodes[right].WinNext = idx
					nodes[right].WinSlot = 2
					byRoundLB[lbr][i] = idx
					idx++
				}
			}
		}
		lbFinalIdx = byRoundLB[lbRoundsGroupDE][0]
	}

	return nodes, wbFinalIdx, lbFinalIdx, nil
}
