package bracket

// StandardSeeds returns the Challonge-style seed ordering for a bracket of
// the given power-of-2 size. The result has `size` elements where position i
// holds the seed number assigned to that slot.
//
// Example outputs:
//
//	size=2  → [1 2]
//	size=4  → [1 4 2 3]
//	size=8  → [1 8 4 5 2 7 3 6]
//	size=16 → [1 16 8 9 4 13 5 12 2 15 7 10 3 14 6 11]
func StandardSeeds(size int) []int {
	if size == 1 {
		return []int{1}
	}
	prev := StandardSeeds(size / 2)
	out := make([]int, 0, size)
	for _, s := range prev {
		out = append(out, s, size+1-s)
	}
	return out
}

// nextPow2 returns the smallest power of two >= n.
func nextPow2(n int) int {
	if n <= 1 {
		return 1
	}
	p := 1
	for p < n {
		p <<= 1
	}
	return p
}

// intLog2 returns log2 of a power-of-2 value n.
func intLog2(n int) int {
	k := 0
	for n > 1 {
		n >>= 1
		k++
	}
	return k
}
