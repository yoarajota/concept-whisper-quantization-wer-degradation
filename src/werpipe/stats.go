package werpipe

import (
	"math"
	"sort"
)

func Aggregate(results []SampleResult) LevelResults {
	lr := LevelResults{Samples: results}
	werList := make([]float64, 0, len(results))
	for _, r := range results {
		if r.Error == nil {
			werList = append(werList, r.WER)
			lr.NumSuccess++
		} else {
			lr.NumError++
		}
	}
	lr.MeanWER = mean(werList)
	lr.MedianWER = median(werList)
	lr.StdDev = stdDev(werList, lr.MeanWER)
	return lr
}

func Compare(baseline, quant []SampleResult) Comparison {
	c := Comparison{
		BaselineWER: mean(sampleWERs(baseline)),
		QuantWER:    mean(sampleWERs(quant)),
	}
	if c.BaselineWER > 0 {
		c.RelChangePct = (c.QuantWER - c.BaselineWER) / c.BaselineWER * 100
	}

	baseWERs := sampleWERs(baseline)
	quantWERs := sampleWERs(quant)
	c.PValue = wilcoxonP(baseWERs, quantWERs)
	c.Significant = c.PValue <= 0.05
	c.BootstrapCI95 = bootstrapCI95(quantWERs)
	return c
}

func sampleWERs(results []SampleResult) []float64 {
	wer := make([]float64, 0, len(results))
	for _, r := range results {
		if r.Error == nil {
			wer = append(wer, r.WER)
		}
	}
	return wer
}

type signedRank struct {
	abs  float64
	rank float64
	sign int
}

func computeSignedRanks(baseline, treatment []float64) []signedRank {
	diffs := make([]signedRank, 0, len(baseline))
	for i := range baseline {
		d := treatment[i] - baseline[i]
		if d == 0 {
			continue
		}
		sign := 1
		if d < 0 {
			sign = -1
		}
		diffs = append(diffs, signedRank{abs: math.Abs(d), sign: sign})
	}
	if len(diffs) == 0 {
		return diffs
	}
	sort.Slice(diffs, func(i, j int) bool { return diffs[i].abs < diffs[j].abs })
	for i := range diffs {
		diffs[i].rank = float64(i + 1)
	}
	averageTiedRanks(diffs)
	return diffs
}

func averageTiedRanks(diffs []signedRank) {
	for j := 0; j < len(diffs); {
		k := j + 1
		for k < len(diffs) && diffs[k].abs == diffs[j].abs {
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
}

func wilcoxonNormalP(diffs []signedRank) float64 {
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
	meanW := nn * (nn + 1) / 4
	varTie := 0.0
	for i := 0; i < len(diffs); {
		k := i + 1
		for k < len(diffs) && diffs[k].abs == diffs[i].abs {
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

func wilcoxonP(baseline, treatment []float64) float64 {
	if len(baseline) == 0 || len(baseline) != len(treatment) {
		return 1.0
	}
	if hasNaN(baseline) || hasNaN(treatment) {
		return 1.0
	}
	diffs := computeSignedRanks(baseline, treatment)
	if len(diffs) == 0 {
		return 1.0
	}
	nn := len(diffs)
	if nn <= 20 {
		var wPos, wNeg float64
		for _, d := range diffs {
			if d.sign > 0 {
				wPos += d.rank
			} else {
				wNeg += d.rank
			}
		}
		return exactWilcoxonP(math.Min(wPos, wNeg), nn)
	}
	return wilcoxonNormalP(diffs)
}

func hasNaN(v []float64) bool {
	for _, x := range v {
		if math.IsNaN(x) {
			return true
		}
	}
	return false
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

func bootstrapCI95(values []float64) [2]float64 {
	if len(values) == 0 {
		return [2]float64{0, 0}
	}
	n := len(values)
	b := 10000
	means := make([]float64, b)
	for i := 0; i < b; i++ {
		sum := 0.0
		for j := 0; j < n; j++ {
			sum += values[fastRandInt(n)]
		}
		means[i] = sum / float64(n)
	}
	sort.Float64s(means)
	lo := int(float64(b) * 0.025)
	hi := int(float64(b) * 0.975)
	return [2]float64{means[lo], means[hi]}
}

func fastRandInt(n int) int {
	return int(fastRand() % uint64(n))
}

var fastRandState uint64 = 1

func fastRand() uint64 {
	fastRandState = fastRandState*6364136223846793005 + 1442695040888963407
	return fastRandState
}

func mean(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range v {
		s += x
	}
	return s / float64(len(v))
}

func median(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	sorted := make([]float64, len(v))
	copy(sorted, v)
	sort.Float64s(sorted)
	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

func stdDev(v []float64, m float64) float64 {
	if len(v) < 2 {
		return 0
	}
	s := 0.0
	for _, x := range v {
		d := x - m
		s += d * d
	}
	return math.Sqrt(s / float64(len(v)-1))
}

func normalCDF(x float64) float64 {
	return 0.5 * math.Erfc(-x / math.Sqrt2)
}
