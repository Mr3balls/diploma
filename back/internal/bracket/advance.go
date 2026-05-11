package bracket

// CascadeReset returns the indices of all matches that must be cleared when
// the match at startIdx is reset. The returned slice includes every match
// reachable via WinNext or LoseNext links from startIdx (not startIdx itself).
//
// Used by tests to verify structural completeness; the service layer does the
// actual DB work using the same graph traversal logic.
func CascadeReset(nodes []Node, startIdx int) []int {
	byIdx := make(map[int]*Node, len(nodes))
	for i := range nodes {
		byIdx[nodes[i].Index] = &nodes[i]
	}

	affected := make([]int, 0)
	visited := make(map[int]bool)

	var walk func(idx int)
	walk = func(idx int) {
		if visited[idx] {
			return
		}
		visited[idx] = true
		nd, ok := byIdx[idx]
		if !ok {
			return
		}
		if nd.WinNext >= 0 {
			affected = append(affected, nd.WinNext)
			walk(nd.WinNext)
		}
		if nd.LoseNext >= 0 {
			affected = append(affected, nd.LoseNext)
			walk(nd.LoseNext)
		}
	}

	walk(startIdx)

	// startIdx was visited but not appended to affected — remove it just in case.
	result := make([]int, 0, len(affected))
	for _, idx := range affected {
		if idx != startIdx {
			result = append(result, idx)
		}
	}
	return result
}
