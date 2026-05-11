package bracket

import "fmt"

// BuildSingle constructs a single-elimination bracket for n participants.
//
// The bracket size S = nextPow2(n). BYE slots are given to the highest seed
// numbers (seeds n+1 … S), so the strongest players get byes when the bracket
// is not a clean power of two.
//
// The returned slice is ordered: WB Round 1, Round 2, …, Final.
// Indices are contiguous 0-based integers; use them to map to real UUIDs.
func BuildSingle(n int) ([]Node, error) {
	if n < 2 {
		return nil, fmt.Errorf("bracket: need at least 2 participants, got %d", n)
	}

	size := nextPow2(n)
	rounds := intLog2(size)
	seeds := StandardSeeds(size)

	nodes := make([]Node, 0, size-1)
	idx := 0

	// byRound[r] holds node indices for round r (1-based).
	byRound := make([][]int, rounds+1)

	// Round 1
	r1count := size / 2
	byRound[1] = make([]int, r1count)
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
			WinSlot:  0,
			LoseNext: -1,
			LoseSlot: 0,
			IsBye:    s1 == 0 || s2 == 0,
		})
		byRound[1][i] = idx
		idx++
	}

	// Rounds 2 … roundsWB
	for r := 2; r <= rounds; r++ {
		prev := byRound[r-1]
		count := len(prev) / 2
		byRound[r] = make([]int, count)
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
				WinSlot:  0,
				LoseNext: -1,
				LoseSlot: 0,
			})
			nodes[left].WinNext = idx
			nodes[left].WinSlot = 1
			nodes[right].WinNext = idx
			nodes[right].WinSlot = 2
			byRound[r][i] = idx
			idx++
		}
	}

	return nodes, nil
}
