package bracket

import "sort"

// GlobalNumbers assigns sequential global match numbers to nodes.
// Ordering: WB (round 1…N, slot 1…M), then LB (round 1…N, slot 1…M), then GF.
// Returns a map from Node.Index → global number (1-based).
//
// The returned numbering is used to populate the global_number column in the
// matches table so the UI can display matches in a canonical order.
func GlobalNumbers(nodes []Node) map[int]int {
	sorted := make([]Node, len(nodes))
	copy(sorted, nodes)

	sectionOrder := map[Section]int{SectionWB: 0, SectionLB: 1, SectionGF: 2}
	sort.Slice(sorted, func(i, j int) bool {
		si := sectionOrder[sorted[i].Section]
		sj := sectionOrder[sorted[j].Section]
		if si != sj {
			return si < sj
		}
		if sorted[i].Round != sorted[j].Round {
			return sorted[i].Round < sorted[j].Round
		}
		return sorted[i].Slot < sorted[j].Slot
	})

	m := make(map[int]int, len(nodes))
	for i, nd := range sorted {
		m[nd.Index] = i + 1
	}
	return m
}
