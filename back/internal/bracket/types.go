// Package bracket contains pure (DB-free) bracket generation and advancement logic.
package bracket

// Section identifies which part of the bracket a match belongs to.
type Section string

const (
	SectionWB Section = "WB" // Winners Bracket
	SectionLB Section = "LB" // Losers Bracket
	SectionGF Section = "GF" // Grand Final
)

// Node represents one match in the bracket, identified by a sequential Index.
// All cross-references use Index values; the service layer maps them to UUIDs.
type Node struct {
	Index   int     // unique 0-based index within this bracket build
	Section Section
	Round   int // round number within Section (1-based)
	Slot    int // position within round (1-based, top-to-bottom)

	// Seed numbers (1-based). 0 means BYE (auto-win).
	// Only set for Round-1 WB matches.
	Seed1 int
	Seed2 int

	// Source matches whose WINNER feeds slot 1 / slot 2 of this match.
	// -1 means the slot is filled from Seed1/Seed2 (no source match).
	Src1 int
	Src2 int

	// WinNext: index of match where this match's WINNER goes. -1 = champion.
	// WinSlot: which slot (1 or 2) in WinNext the winner occupies.
	WinNext int
	WinSlot int

	// LoseNext: index of match where this match's LOSER goes (WB→LB only).
	// -1 = loser is eliminated (LB matches, GF).
	// LoseSlot: which slot (1 or 2) in LoseNext the loser occupies.
	LoseNext int
	LoseSlot int

	// IsBye is true when one side is a BYE (auto-advance, no real match).
	IsBye bool
}
