package werpipe

import (
	"math"
	"testing"
)

func TestSmallNNoSignificance(t *testing.T) {
	n := 5
	base := make([]SampleResult, n)
	quant := make([]SampleResult, n)
	for i := 0; i < n; i++ {
		base[i].WER = 0.05
		quant[i].WER = 0.051
	}
	c := Compare(base, quant)
	if c.Significant {
		t.Errorf("small N with tiny shift should not be significant, got p=%f", c.PValue)
	}
}

func TestWilcoxonOnNaN(t *testing.T) {
	p := wilcoxonP([]float64{math.NaN(), 0.1}, []float64{0.2, 0.3})
	if p != 1.0 {
		t.Errorf("NaN should default to non-significant, got p=%f", p)
	}
}

func TestWilcoxonOnSinglePair(t *testing.T) {
	p := wilcoxonP([]float64{0.1}, []float64{0.2})
	if p <= 0.05 {
		t.Errorf("single pair should never be significant, got p=%f", p)
	}
}
