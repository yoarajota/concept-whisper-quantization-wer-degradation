package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/yoarajota/concept-whisper-quantization-wer-degradation/src/werpipe"
)

func TestMergeChunks(t *testing.T) {
	chunk1 := []levelReport{
		{
			Level: "f16", Model: "a.bin", NumSamples: 2, NumErrors: 0,
			Results: werpipe.LevelResults{Samples: []werpipe.SampleResult{
				{SampleID: "s1", WER: 0.10},
				{SampleID: "s2", WER: 0.20},
			}},
		},
		{
			Level: "q4_0", Model: "b.bin", NumSamples: 2, NumErrors: 0,
			Results: werpipe.LevelResults{Samples: []werpipe.SampleResult{
				{SampleID: "s1", WER: 0.12},
				{SampleID: "s2", WER: 0.24},
			}},
		},
	}
	chunk2 := []levelReport{
		{
			Level: "f16", Model: "a.bin", NumSamples: 2, NumErrors: 0,
			Results: werpipe.LevelResults{Samples: []werpipe.SampleResult{
				{SampleID: "s3", WER: 0.30},
				{SampleID: "s4", WER: 0.40},
			}},
		},
		{
			Level: "q4_0", Model: "b.bin", NumSamples: 2, NumErrors: 0,
			Results: werpipe.LevelResults{Samples: []werpipe.SampleResult{
				{SampleID: "s3", WER: 0.36},
				{SampleID: "s4", WER: 0.48},
			}},
		},
	}

	dir := t.TempDir()
	f1 := dir + "/c1.json"
	f2 := dir + "/c2.json"
	if err := writeChunk(f1, chunk1); err != nil {
		t.Fatal(err)
	}
	if err := writeChunk(f2, chunk2); err != nil {
		t.Fatal(err)
	}

	oldStdout := os.Stdout
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = devNull
	defer func() { os.Stdout = oldStdout }()

	runMerge([]string{f1, f2})
}

func writeChunk(path string, reports []levelReport) error {
	data, err := json.Marshal(reports)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
