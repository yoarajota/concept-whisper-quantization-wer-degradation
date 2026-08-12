package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yoarajota/concept-whisper-quantization-wer-degradation/src/werpipe"
)

func main() {
	audioDir := flag.String("audio", "", "path to LibriSpeech test-clean audio directory")
	transDir := flag.String("transcripts", "", "path to LibriSpeech test-clean transcripts")
	whisperCLI := flag.String("whisper-cli", "/usr/local/bin/whisper-cli", "path to whisper-cli binary")
	modelDir := flag.String("model-dir", "/models", "directory containing quantized ggml model files")
	threads := flag.Int("threads", 4, "number of CPU threads")
	levelsFlag := flag.String("levels", "f16,q8_0,q5_0,q4_0", "comma-separated quantization levels")
	verbose := flag.Bool("v", false, "verbose: print per-sample WER")
	flag.Parse()

	if *audioDir == "" || *transDir == "" {
		fmt.Fprintf(os.Stderr, "usage: werpipe -audio <dir> -transcripts <dir>\n")
		os.Exit(1)
	}

	modelMap := map[string]string{
		"f16":  "ggml-large-v3.bin",
		"q8_0": "ggml-large-v3-q8_0.bin",
		"q5_0": "ggml-large-v3-q5_0.bin",
		"q4_0": "ggml-large-v3-q4_0.bin",
	}

	levels := strings.Split(*levelsFlag, ",")
	samples := loadSamples(*audioDir, *transDir)
	fmt.Fprintf(os.Stderr, "loaded %d samples\n", len(samples))
	if len(samples) == 0 {
		fmt.Fprintln(os.Stderr, "no samples found — check audio/transcript paths")
		os.Exit(1)
	}

	pipeline := werpipe.NewPipeline(*whisperCLI, *modelDir, *threads)

	type levelReport struct {
		Level     string              `json:"level"`
		Model     string              `json:"model"`
		NumSamples int                `json:"num_samples"`
		NumErrors int                `json:"num_errors"`
		Results   werpipe.LevelResults `json:"results"`
		Comparison *werpipe.Comparison  `json:"comparison,omitempty"`
	}
	var reports []levelReport
	var baselineResults []werpipe.SampleResult

	for i, lvl := range levels {
		model, ok := modelMap[lvl]
		if !ok {
			model = lvl
		}
		modelPath := filepath.Join(*modelDir, model)
		if _, err := os.Stat(modelPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "SKIP %s: model not found at %s\n", lvl, modelPath)
			continue
		}

		results, err := pipeline.Run(werpipe.QuantLevel(lvl), model, samples)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: run error: %v\n", lvl, err)
			continue
		}

		agg := werpipe.Aggregate(results)
		fmt.Fprintf(os.Stderr, "%s: WER=%.4f (median=%.4f, std=%.4f, %d ok/%d errors)\n",
			lvl, agg.MeanWER, agg.MedianWER, agg.StdDev, agg.NumSuccess, agg.NumError)

		r := levelReport{
			Level:      lvl,
			Model:      model,
			NumSamples: len(samples),
			NumErrors:  agg.NumError,
			Results:    agg,
		}

		if i == 0 {
			baselineResults = results
		} else if baselineResults != nil {
			cmp := werpipe.Compare(baselineResults, results)
			r.Comparison = &cmp
			fmt.Fprintf(os.Stderr, "  vs f16: rel=%+.1f%%, p=%.4f, sig=%v, 95%% CI=[%.4f, %.4f]\n",
				cmp.RelChangePct, cmp.PValue, cmp.Significant, cmp.BootstrapCI95[0], cmp.BootstrapCI95[1])
		}

		if *verbose {
			for _, s := range results {
				if s.Error != nil {
					fmt.Fprintf(os.Stderr, "  %s: error %v\n", s.SampleID, s.Error)
				} else {
					fmt.Fprintf(os.Stderr, "  %s: WER=%.4f\n", s.SampleID, s.WER)
				}
			}
		}

		reports = append(reports, r)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(reports)
}

func loadSamples(audioDir, transDir string) []werpipe.Sample {
	var samples []werpipe.Sample
	filepath.Walk(audioDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".flac" && ext != ".wav" {
			return nil
		}
		rel, _ := filepath.Rel(audioDir, path)
		id := strings.TrimSuffix(rel, filepath.Ext(rel))
		transPath := filepath.Join(transDir, id+".txt")
		txt, err := os.ReadFile(transPath)
		if err != nil {
			return nil
		}
		samples = append(samples, werpipe.Sample{
			ID:        id,
			AudioPath: path,
			Reference: strings.TrimSpace(string(txt)),
		})
		return nil
	})
	return samples
}
