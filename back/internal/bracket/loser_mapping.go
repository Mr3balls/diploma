package bracket

// MapLoserToLB returns the 0-based LB match slot and slot-in-match (1 or 2)
// for a WB loser dropping into the Losers Bracket.
//
// Parameters (all 0-based):
//   wbRound     – WB round number (0-based: 0 = WBR1, 1 = WBR2 …)
//   wbSlot      – slot within that WB round (0-based)
//   lbMatchCount – number of matches in the target LB round
//
// WBR1 losers → LBR1 (minor): fold/reverse pairing avoids immediate rematches.
// WBR(k+1) losers → LBR(2k) (major): alternate reverse/natural to separate
// players who have already met.
//
// This mapping has been structurally verified for n ≤ 32 by
// TestWBR1OpponentsNotPairedInLBR1 and TestNoRematchBeforeGrandFinal.
// Full "no rematch anywhere before GF" is not mathematically guaranteed in
// double elimination with arbitrary results; verify for n > 32 before extending.
func MapLoserToLB(wbRound, wbSlot, lbMatchCount int) (lbSlot int, slotInMatch int) {
	if wbRound == 0 {
		// WBR1 → LBR1 fold pairing
		// Lower half (wbSlot < lbMatchCount) → slot 1
		// Upper half (wbSlot ≥ lbMatchCount) → slot 2, mirrored index
		if wbSlot < lbMatchCount {
			return wbSlot, 1
		}
		return 2*lbMatchCount - 1 - wbSlot, 2
	}
	// WBR(k+1) → LBR(2k) major round, always slot 2
	// Odd k → reverse order to reduce rematches
	k := wbRound // wbRound here is already k (0-based from WBR2)
	if k%2 == 1 {
		return lbMatchCount - 1 - wbSlot, 2
	}
	return wbSlot, 2
}
