package werpipe

import "strings"

func ComputeWER(reference, hypothesis string) float64 {
	ref := strings.Fields(reference)
	hyp := strings.Fields(hypothesis)
	n := len(ref)
	if n == 0 {
		if len(hyp) == 0 {
			return 0
		}
		return float64(len(hyp))
	}
	d := levenshteinDistance(ref, hyp)
	return float64(d) / float64(n)
}

func levenshteinDistance(a, b []string) int {
	m, n := len(a), len(b)
	prev := make([]int, n+1)
	curr := make([]int, n+1)
	for j := 0; j <= n; j++ {
		prev[j] = j
	}
	for i := 1; i <= m; i++ {
		curr[0] = i
		for j := 1; j <= n; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min3(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[n]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
