package service

import "testing"

func TestNextPowerOfTwo(t *testing.T) {
	cases := map[int]int{1: 1, 2: 2, 3: 4, 5: 8, 8: 8, 9: 16}
	for in, want := range cases {
		if got := nextPowerOfTwo(in); got != want {
			t.Fatalf("nextPowerOfTwo(%d)=%d want %d", in, got, want)
		}
	}
}

func TestBuildBracketMatchesCreatesFullTree(t *testing.T) {
	teamIDs := []string{"t1", "t2", "t3", "t4", "t5"}
	matches, err := buildBracketMatches("tour-1", "bracket-1", teamIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 7 {
		t.Fatalf("expected 7 matches for 8-slot single elimination tree, got %d", len(matches))
	}
	if matches[0].RoundNumber != 1 {
		t.Fatalf("expected first match to be in round 1")
	}
	propagateByes(matches)
	byeWinners := 0
	for _, match := range matches {
		if match.IsBye && match.WinnerTeamID != nil {
			byeWinners++
		}
	}
	if byeWinners == 0 {
		t.Fatalf("expected byes to be propagated for non-power-of-two team count")
	}
}
