package main

import "strings"

func werScore(reference, hypothesis string) float64 {
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
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
		dp[i][0] = i
	}
	for j := 0; j <= n; j++ {
		dp[0][j] = j
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			dp[i][j] = min3(
				dp[i-1][j]+1,
				dp[i][j-1]+1,
				dp[i-1][j-1]+cost,
			)
		}
	}
	return dp[m][n]
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
