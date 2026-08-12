package main

import (
	"math"
	"sort"
)

type rankedDiff struct {
	absDiff float64
	rank    float64
	sign    int
}

func wilcoxon(baseline, treatment []float64) float64 {
	n := len(baseline)
	if n == 0 || len(treatment) != n {
		return 1.0
	}

	diffs := make([]rankedDiff, 0, n)
	for i := range baseline {
		d := treatment[i] - baseline[i]
		if d == 0 {
			continue
		}
		sign := 1
		if d < 0 {
			sign = -1
		}
		diffs = append(diffs, rankedDiff{absDiff: math.Abs(d), sign: sign})
	}

	if len(diffs) == 0 {
		return 1.0
	}

	sort.Slice(diffs, func(i, j int) bool {
		return diffs[i].absDiff < diffs[j].absDiff
	})

	for i := range diffs {
		diffs[i].rank = float64(i + 1)
	}

	j := 0
	for j < len(diffs) {
		k := j + 1
		for k < len(diffs) && diffs[k].absDiff == diffs[j].absDiff {
			k++
		}
		if k > j+1 {
			sum := 0.0
			for m := j; m < k; m++ {
				sum += diffs[m].rank
			}
			avg := sum / float64(k-j)
			for m := j; m < k; m++ {
				diffs[m].rank = avg
			}
		}
		j = k
	}

	var wPos, wNeg float64
	for _, d := range diffs {
		if d.sign > 0 {
			wPos += d.rank
		} else {
			wNeg += d.rank
		}
	}

	nn := float64(len(diffs))
	w := math.Min(wPos, wNeg)

	if nn <= 20 {
		return exactWilcoxonP(w, int(nn))
	}

	meanW := nn * (nn + 1) / 4
	varTie := .0
	for i := 0; i < len(diffs); {
		k := i + 1
		for k < len(diffs) && diffs[k].absDiff == diffs[i].absDiff {
			k++
		}
		t := float64(k - i)
		varTie += t*t*t - t
		i = k
	}
	stdW := math.Sqrt(nn*(nn+1)*(2*nn+1)/24 - varTie/48)
	z := (w - meanW) / stdW
	return 2 * normalCDF(-math.Abs(z))
}

func exactWilcoxonP(w float64, n int) float64 {
	if n > 20 {
		return 1.0
	}
	maxRankSum := float64(n*(n+1)) / 2
	count := 0
	total := 0.0
	var recurse func(int, float64)
	recurse = func(i int, sum float64) {
		if i > n {
			total += 1
			if sum <= w || (maxRankSum-sum) <= w {
				count++
			}
			return
		}
		recurse(i+1, sum)
		recurse(i+1, sum+float64(i))
	}
	recurse(1, 0)
	p := float64(count) / total
	if p > 1.0 {
		p = 1.0
	}
	return p
}

func normalCDF(x float64) float64 {
	return 0.5 * math.Erfc(-x/math.Sqrt2)
}
