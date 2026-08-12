package werpipe

import (
	"math"
	"strings"
	"testing"
)

func TestComputeWERPerfect(t *testing.T) {
	w := ComputeWER("hello world", "hello world")
	if w != 0 {
		t.Errorf("perfect WER = %f, want 0", w)
	}
}

func TestComputeWERHalfWrong(t *testing.T) {
	w := ComputeWER("hello world", "hello")
	if w != 0.5 {
		t.Errorf("half wrong WER = %f, want 0.5", w)
	}
}

func TestComputeWERAllWrong(t *testing.T) {
	w := ComputeWER("hello world", "completely different text today")
	if w != 2.0 {
		t.Errorf("all wrong WER = %f, want 2.0", w)
	}
}

func TestComputeWEREmpty(t *testing.T) {
	if ComputeWER("", "") != 0 {
		t.Error("empty strings should give 0")
	}
	if ComputeWER("hello", "") != 1.0 {
		t.Error("only ref should give 1.0")
	}
	if ComputeWER("", "hello") != 1.0 {
		t.Error("only hyp should give 1.0")
	}
}

func TestNormalize(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Hello, World!", "hello world"},
		{"  MULTIPLE   SPACES  ", "multiple spaces"},
		{"Numbers 123 test", "numbers 123 test"},
		{"It's a test", "it's a test"},
		{"Punctuation... removed!!!", "punctuation removed"},
		{"UPPERCASE", "uppercase"},
	}
	for _, c := range cases {
		got := Normalize(c.in)
		if got != c.want {
			t.Errorf("Normalize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestAggregateEmpty(t *testing.T) {
	lr := Aggregate([]SampleResult{{SampleID: "a", Error: errTest}})
	if lr.NumError != 1 || lr.NumSuccess != 0 {
		t.Error("empty results should have errors counted")
	}
}

var errTest = &testError{}

type testError struct{}

func (e *testError) Error() string { return "test error" }

func TestAggregate(t *testing.T) {
	results := []SampleResult{
		{WER: 0.1},
		{WER: 0.2},
		{WER: 0.3},
	}
	lr := Aggregate(results)
	if lr.NumSuccess != 3 {
		t.Errorf("NumSuccess = %d, want 3", lr.NumSuccess)
	}
	if math.Abs(lr.MeanWER-0.2) > 1e-9 {
		t.Errorf("MeanWER = %f, want 0.2", lr.MeanWER)
	}
}

func TestCompareIdentical(t *testing.T) {
	base := []SampleResult{{WER: 0.1}, {WER: 0.2}, {WER: 0.15}}
	c := Compare(base, base)
	if c.PValue != 1.0 || c.Significant {
		t.Errorf("identical data should give p=1.0, got p=%f sig=%v", c.PValue, c.Significant)
	}
	if c.RelChangePct != 0 {
		t.Errorf("identical data should give 0 rel change, got %f", c.RelChangePct)
	}
}

func TestWilcoxonSystematicShift(t *testing.T) {
	n := 40
	base := make([]SampleResult, n)
	quant := make([]SampleResult, n)
	for i := 0; i < n; i++ {
		base[i].WER = 0.05
		quant[i].WER = 0.10
	}
	c := Compare(base, quant)
	if c.PValue > 0.05 || !c.Significant {
		t.Errorf("systematic shift: p=%f, sig=%v — should be significant", c.PValue, c.Significant)
	}
}

func TestBootstrapCI(t *testing.T) {
	values := []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	ci := bootstrapCI95(values)
	if ci[0] < 1 || ci[1] > 8 {
		t.Errorf("95%% CI should be [~1, ~8], got [%f, %f]", ci[0], ci[1])
	}
}

func TestSampleWERsSkipsErrors(t *testing.T) {
	results := []SampleResult{
		{WER: 0.1},
		{Error: errTest, WER: math.NaN()},
		{WER: 0.3},
	}
	wer := sampleWERs(results)
	if len(wer) != 2 {
		t.Errorf("expected 2 valid WERs, got %d", len(wer))
	}
}

func TestLevenshteinWords(t *testing.T) {
	a := strings.Fields("the quick brown fox")
	b := strings.Fields("the slow brown fox")
	d := levenshteinDistance(a, b)
	if d != 1 {
		t.Errorf("levenshtein = %d, want 1", d)
	}
}
