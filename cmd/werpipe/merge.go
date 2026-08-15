package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/yoarajota/whisper-quantization-wer-degradation/src/werpipe"
)

type levelReport struct {
	Level      string               `json:"level"`
	Model      string               `json:"model"`
	NumSamples int                  `json:"num_samples"`
	NumErrors  int                  `json:"num_errors"`
	Results    werpipe.LevelResults `json:"results"`
	Comparison *werpipe.Comparison  `json:"comparison,omitempty"`
}

func runMerge(files []string) {
	if len(files) < 2 {
		fmt.Fprintf(os.Stderr, "usage: werpipe merge chunk1.json chunk2.json [...]\n")
		os.Exit(1)
	}

	var reports [][]levelReport
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", f, err)
			os.Exit(1)
		}
		var chunk []levelReport
		if err := json.Unmarshal(data, &chunk); err != nil {
			fmt.Fprintf(os.Stderr, "parse %s: %v\n", f, err)
			os.Exit(1)
		}
		if len(chunk) == 0 {
			fmt.Fprintf(os.Stderr, "%s: empty chunk\n", f)
			os.Exit(1)
		}
		reports = append(reports, chunk)
	}

	first := reports[0]
	for _, chunk := range reports[1:] {
		if len(chunk) != len(first) {
			fmt.Fprintf(os.Stderr, "chunk level count mismatch: %d vs %d\n", len(first), len(chunk))
			os.Exit(1)
		}
		for i := range first {
			if chunk[i].Level != first[i].Level {
				fmt.Fprintf(os.Stderr, "chunk level order mismatch at %d: %s vs %s\n",
					i, chunk[i].Level, first[i].Level)
				os.Exit(1)
			}
		}
	}

	var combined []levelReport
	var baselineResults []werpipe.SampleResult
	for li := range first {
		merged := make([]werpipe.SampleResult, 0)
		for _, chunk := range reports {
			merged = append(merged, chunk[li].Results.Samples...)
		}
		sort.Slice(merged, func(i, j int) bool { return merged[i].SampleID < merged[j].SampleID })

		agg := werpipe.Aggregate(merged)
		r := levelReport{
			Level:      first[li].Level,
			Model:      first[li].Model,
			NumSamples: len(merged),
			NumErrors:  agg.NumError,
			Results:    agg,
		}
		if li == 0 {
			baselineResults = merged
		} else if baselineResults != nil {
			cmp := werpipe.Compare(baselineResults, merged)
			r.Comparison = &cmp
			fmt.Fprintf(os.Stderr, "%-30s WER=%.4f rel=%+.1f%% p=%.4f sig=%v 95%% CI=[%.4f, %.4f] (n=%d)\n",
				first[li].Level, agg.MeanWER, cmp.RelChangePct, cmp.PValue,
				cmp.Significant, cmp.BootstrapCI95[0], cmp.BootstrapCI95[1], agg.NumSuccess)
		} else {
			fmt.Fprintf(os.Stderr, "%-30s WER=%.4f (n=%d)\n", first[li].Level, agg.MeanWER, agg.NumSuccess)
		}
		combined = append(combined, r)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(combined); err != nil {
		fmt.Fprintf(os.Stderr, "encode: %v\n", err)
		os.Exit(1)
	}
}
