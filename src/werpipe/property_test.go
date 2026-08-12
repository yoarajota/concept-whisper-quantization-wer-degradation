package werpipe

import (
	"testing"
)

func TestWERBoundedBelow(t *testing.T) {
	for _, tt := range []struct{ ref, hyp string }{
		{"a", "a a a a a a a a a a"},
		{"the quick brown fox", "and now for something completely different and entirely unrelated"},
	} {
		w := ComputeWER(tt.ref, tt.hyp)
		if w < 0 {
			t.Errorf("WER(%q, %q) = %f, WER must be >= 0", tt.ref, tt.hyp, w)
		}
	}
}

func TestNormalizeIdempotent(t *testing.T) {
	inputs := []string{"hello world", "UPPERCASE", "  spaces  ", "it's"}
	for _, in := range inputs {
		once := Normalize(in)
		twice := Normalize(once)
		if once != twice {
			t.Errorf("Normalize not idempotent: %q -> %q -> %q", in, once, twice)
		}
	}
}

func TestNormalizeLosesNoWords(t *testing.T) {
	in := "Hello, world! 123 test."
	out := Normalize(in)
	expectedWords := 4
	if len(words(out)) != expectedWords {
		t.Errorf("Normalize(%q) produced %d words, expected %d", in, len(words(out)), expectedWords)
	}
}

func words(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	start := -1
	for i, r := range s {
		if r == ' ' {
			if start >= 0 {
				result = append(result, s[start:i])
				start = -1
			}
		} else if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		result = append(result, s[start:])
	}
	return result
}

func TestMeanEmpty(t *testing.T) {
	if mean(nil) != 0 || mean([]float64{}) != 0 {
		t.Error("mean of empty slice should be 0")
	}
}

func TestMedianOddEven(t *testing.T) {
	if median([]float64{3, 1, 2}) != 2 {
		t.Error("median of [1,2,3] should be 2")
	}
	if median([]float64{4, 1, 2, 3}) != 2.5 {
		t.Error("median of [1,2,3,4] should be 2.5")
	}
}

func TestStdDev(t *testing.T) {
	v := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	s := stdDev(v, 5)
	if s < 2 || s > 2.5 {
		t.Errorf("stdDev = %f, expected ~2.1", s)
	}
}

func TestBootstrapCIMonotonic(t *testing.T) {
	values := make([]float64, 100)
	for i := range values {
		values[i] = float64(i) / 100
	}
	ci := bootstrapCI95(values)
	if ci[0] > ci[1] {
		t.Errorf("CI lower > upper: [%f, %f]", ci[0], ci[1])
	}
	if ci[0] < 0 || ci[1] > 1 {
		t.Errorf("CI out of range: [%f, %f]", ci[0], ci[1])
	}
}
