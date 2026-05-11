package bracket

import "fmt"

// BuildDouble constructs a double-elimination bracket for n participants.
//
// Structure for bracket size S = nextPow2(n), roundsWB = log2(S):
//
//	WB:  standard single-elimination on S slots (S-1 matches)
//	LB:  2*(roundsWB-1) rounds alternating minor/major (S-2 matches total)
//	GF:  1 match (WB champion vs LB champion)
//
// LB round pattern:
//
//	LBR1  (minor) – WBR1 losers pair up (fold/reverse pairing)
//	LBR2  (major) – LBR1 winners vs WBR2 losers
//	LBR3  (minor) – LBR2 winners pair up
//	LBR4  (major) – LBR3 winners vs WBR3 losers
//	…
//	LBR(2k-1) (minor)
//	LBR(2k)   (major) – vs WBR(k+1) losers
//
// WB→LB loser mapping alternates reverse/identity to separate initial opponents:
//   - odd k  → WB losers inserted in reversed slot order
//   - even k → WB losers inserted in natural slot order
//
// GF: Team1 = WB champion (slot 1), Team2 = LB champion (slot 2).
// If Team2 wins, the caller must create a Grand Final Reset match (round 2, same GF).
//
// For n=2 the function degenerates to a single GF match (no LB).
func BuildDouble(n int) ([]Node, error) {
	if n < 2 {
		return nil, fmt.Errorf("bracket: need at least 2 participants, got %d", n)
	}

	size := nextPow2(n)
	roundsWB := intLog2(size)
	seeds := StandardSeeds(size)

	// Upper bound on total nodes: WB(size-1) + LB(size-2) + GF(1)
	nodes := make([]Node, 0, 2*size)
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

	// ── LB (only when size >= 4) ──────────────────────────────────────────
	lbRounds := 2 * (roundsWB - 1) // 0 when roundsWB==1 (size==2)
	byRoundLB := make([][]int, lbRounds+1)

	if lbRounds > 0 {
		// LBR1 (minor): fold WBR1 losers
		// LBR1[i] pairs WBR1[i] loser vs WBR1[r1count-1-i] loser.
		// Only iterate i < r1count/2 to avoid duplicate pairings.
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
				Src1:     -1, // filled from WB losers via LoseNext, not winner-source
				Src2:     -1,
				WinNext:  -1,
				LoseNext: -1,
			})
			// WB match losers drop here
			nodes[wbLeft].LoseNext = idx
			nodes[wbLeft].LoseSlot = 1
			nodes[wbRight].LoseNext = idx
			nodes[wbRight].LoseSlot = 2
			byRoundLB[1][i] = idx
			idx++
		}

		// LBR2…lbRounds
		for lbr := 2; lbr <= lbRounds; lbr++ {
			isMajor := lbr%2 == 0

			if isMajor {
				// Major round: LB winners from previous round vs WB losers.
				// LBR(2k) receives losers from WBR(k+1).
				k := lbr / 2
				wbr := k + 1
				wbLosers := byRoundWB[wbr]
				lbPrev := byRoundLB[lbr-1]
				count := len(lbPrev) // == len(wbLosers)

				byRoundLB[lbr] = make([]int, count)
				for i := 0; i < count; i++ {
					// Alternate reverse/identity to minimise rematches:
					// odd k → reversed WB loser order
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
						Src1:     lbWinnerIdx, // LB winner feeds slot 1
						Src2:     -1,          // WB loser feeds slot 2 via LoseNext
						WinNext:  -1,
						LoseNext: -1,
					})
					nodes[lbWinnerIdx].WinNext = idx
					nodes[lbWinnerIdx].WinSlot = 1
					// WB match loser drops here into slot 2
					nodes[wbLoserIdx].LoseNext = idx
					nodes[wbLoserIdx].LoseSlot = 2
					byRoundLB[lbr][i] = idx
					idx++
				}
			} else {
				// Minor round: pair up previous LB round's winners.
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
	}

	// ── Grand Final ───────────────────────────────────────────────────────
	wbFinalIdx := byRoundWB[roundsWB][0]
	gfIdx := idx

	if lbRounds > 0 {
		lbFinalIdx := byRoundLB[lbRounds][0]
		nodes = append(nodes, Node{
			Index:    gfIdx,
			Section:  SectionGF,
			Round:    1,
			Slot:     1,
			Src1:     wbFinalIdx, // WB champion → slot 1
			Src2:     lbFinalIdx, // LB champion → slot 2
			WinNext:  -1,
			LoseNext: -1,
		})
		nodes[wbFinalIdx].WinNext = gfIdx
		nodes[wbFinalIdx].WinSlot = 1
		nodes[lbFinalIdx].WinNext = gfIdx
		nodes[lbFinalIdx].WinSlot = 2
	} else {
		// n=2: one WB match + one GF match.
		// WB winner → GF slot 1 (via WinNext).
		// WB loser  → GF slot 2 directly (via LoseNext, no real LB).
		gfIdx2 := idx
		nodes = append(nodes, Node{
			Index:    gfIdx2,
			Section:  SectionGF,
			Round:    1,
			Slot:     1,
			Src1:     wbFinalIdx, // WB winner feeds slot 1
			Src2:     -1,         // WB loser feeds slot 2 via LoseNext
			WinNext:  -1,
			LoseNext: -1,
		})
		nodes[wbFinalIdx].WinNext = gfIdx2
		nodes[wbFinalIdx].WinSlot = 1
		nodes[wbFinalIdx].LoseNext = gfIdx2
		nodes[wbFinalIdx].LoseSlot = 2
	}

	return nodes, nil
}
