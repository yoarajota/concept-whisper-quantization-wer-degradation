package main

import (
	"math"
	"strings"
	"testing"
)

func TestWerScore(t *testing.T) {
	cases := []struct {
		ref, hyp string
		expected float64
	}{
		{"hello world", "hello world", 0.0},
		{"hello world", "hello", 0.5},
		{"hello world", "hello world today", 0.5},
		{"hello world", "hi world", 0.5},
		{"", "", 0.0},
		{"hello", "", 1.0},
		{"", "hello", 1.0},
		{"the quick brown fox", "the quick brown dog", 0.25},
	}
	for _, c := range cases {
		got := werScore(c.ref, c.hyp)
		if math.Abs(got-c.expected) > 1e-9 {
			t.Errorf("werScore(%q, %q) = %f, want %f", c.ref, c.hyp, got, c.expected)
		}
	}
}

func TestLevenshtein(t *testing.T) {
	d := levenshteinDistance(strings.Split("kitten", ""), strings.Split("sitting", ""))
	if d != 3 {
		t.Errorf("levenshtein(kitten, sitting) = %d, want 3", d)
	}
}

func TestWilcoxonIdentical(t *testing.T) {
	a := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	p := wilcoxon(a, a)
	if p != 1.0 {
		t.Errorf("identical data should give p=1.0, got %f", p)
	}
}

func TestWilcoxonShift(t *testing.T) {
	n := 60
	baseline := make([]float64, n)
	treatment := make([]float64, n)
	for i := 0; i < n; i++ {
		baseline[i] = 0.0
		treatment[i] = 0.01
	}
	p := wilcoxon(baseline, treatment)
	if p > 0.05 {
		t.Errorf("systematic shift should give p <= 0.05, got %f", p)
	}
}

func TestLevenshteinWords(t *testing.T) {
	a := []string{"the", "quick", "brown", "fox"}
	b := []string{"the", "slow", "brown", "fox"}
	d := levenshteinDistance(a, b)
	if d != 1 {
		t.Errorf("levenshtein one substitution = %d, want 1", d)
	}
}

func TestWerScorePerfect(t *testing.T) {
	ref := "the quick brown fox jumps over the lazy dog"
	hyp := "the quick brown fox jumps over the lazy dog"
	w := werScore(ref, hyp)
	if w != 0.0 {
		t.Errorf("perfect match WER = %f, want 0.0", w)
	}
}

func TestWerScoreAllWrong(t *testing.T) {
	ref := "hello world"
	hyp := "completely different text today"
	w := werScore(ref, hyp)
	if w != 2.0 {
		t.Errorf("all wrong WER = %f, want 2.0 (2 subs + 2 ins / 2 ref)", w)
	}
}
