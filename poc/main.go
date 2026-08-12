package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type quantLevel struct {
	name  string
	model string
}

type sample struct {
	audioPath string
	reference string
	hypotheses map[string]string
	wer        map[string]float64
}

func main() {
	audioDir := flagOrEnv("AUDIO_DIR", os.Getenv("AUDIO_DIR"))
	transDir := flagOrEnv("TRANS_DIR", os.Getenv("TRANS_DIR"))
	whisperCppDir := flagOrEnv("WHISPER_DIR", os.Getenv("WHISPER_DIR"))
	if audioDir == "" || transDir == "" || whisperCppDir == "" {
		fmt.Fprintf(os.Stderr, "usage: AUDIO_DIR=/path/to/LibriSpeech/test-clean TRANS_DIR=/path/to/transcripts WHISPER_DIR=/path/to/whisper.cpp go run .\n")
		os.Exit(1)
	}

	cli := filepath.Join(whisperCppDir, "build", "bin", "whisper-cli")
	if _, err := os.Stat(cli); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "whisper-cli not found at %s — build whisper.cpp first:\n  cmake -B build && cmake --build build -j --config Release\n", cli)
		os.Exit(1)
	}

	levels := []quantLevel{
		{name: "f16", model: "models/ggml-large-v3.bin"},
		{name: "q8_0", model: "models/ggml-large-v3-q8_0.bin"},
		{name: "q5_0", model: "models/ggml-large-v3-q5_0.bin"},
		{name: "q4_0", model: "models/ggml-large-v3-q4_0.bin"},
	}

	samples, err := loadSamples(audioDir, transDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load samples: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("loaded %d samples\n", len(samples))

	for _, lvl := range levels {
		runLevel(cli, lvl, samples)
	}

	report(samples, levels)
}

func runLevel(cli string, lvl quantLevel, samples []sample) {
	modelPath := filepath.Join(filepath.Dir(cli), "..", "..", lvl.model)
	fmt.Printf("\n%s (%s) ", lvl.name, lvl.model)
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		fmt.Printf("SKIP — model not found\n  download: cd whisper.cpp && sh models/download-ggml-model.sh large-v3\n  quantize: ./build/bin/quantize models/ggml-large-v3.bin models/ggml-large-v3-%s.bin %s\n", lvl.name, lvl.name)
		return
	}
	fmt.Println()

	for i := range samples {
		out, err := transcribe(cli, modelPath, samples[i].audioPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [%d] transcribe error: %v\n", i, err)
			continue
		}
		hyp := normalize(out)
		samples[i].hypotheses[lvl.name] = hyp
		samples[i].wer[lvl.name] = werScore(samples[i].reference, hyp)
	}
}

func report(samples []sample, levels []quantLevel) {
	fmt.Println("\n=== Results ===")
	fmt.Printf("%-6s | %10s | %10s | %10s\n", " level", "  WER", "rel chg", "median")
	fmt.Println(strings.Repeat("-", 50))

	var baselineWER []float64
	perLevel := map[string][]float64{}

	for _, lvl := range levels {
		werList := make([]float64, 0, len(samples))
		for _, s := range samples {
			if w, ok := s.wer[lvl.name]; ok {
				werList = append(werList, w)
			}
		}
		perLevel[lvl.name] = werList
		if lvl.name == "f16" {
			baselineWER = werList
		}
	}

	sort.Float64s(baselineWER)
	baseWER := mean(baselineWER)

	for _, lvl := range levels {
		werList := perLevel[lvl.name]
		if len(werList) == 0 {
			continue
		}
		avg := mean(werList)
		rel := 0.0
		if baseWER > 0 {
			rel = (avg - baseWER) / baseWER * 100
		}
		m := median(werList)
		sig := ""
		if lvl.name != "f16" {
			p := wilcoxon(baselineWER, werList)
			if p <= 0.05 {
				sig = fmt.Sprintf("  p=%.4f *", p)
			} else {
				sig = fmt.Sprintf("  p=%.4f", p)
			}
		} else {
			sig = "  (baseline)"
		}
		fmt.Printf("%-6s | %10.4f | %+9.1f%% | %10.4f |%s\n", lvl.name, avg, rel, m, sig)
	}

	if len(baselineWER) == 0 {
		fmt.Println("\nno baseline (f16) data — check model path")
	}
}

func loadSamples(audioDir, transDir string) ([]sample, error) {
	var samples []sample
	err := filepath.Walk(audioDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".flac") && !strings.HasSuffix(path, ".wav") {
			return nil
		}
		rel, _ := filepath.Rel(audioDir, path)
		rel = strings.TrimSuffix(rel, filepath.Ext(rel))
		transPath := filepath.Join(transDir, rel+".txt")
		txt, err := os.ReadFile(transPath)
		if err != nil {
			return nil
		}
		ref := normalize(string(txt))
		samples = append(samples, sample{
			audioPath:  path,
			reference:  ref,
			hypotheses: map[string]string{},
			wer:        map[string]float64{},
		})
		return nil
	})
	return samples, err
}

func transcribe(cli, modelPath, audioPath string) (string, error) {
	cmd := exec.Command(cli,
		"-m", modelPath,
		"-f", audioPath,
		"--no-timestamps",
		"-l", "en",
		"-t", "4",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

func normalize(text string) string {
	text = strings.ToLower(text)
	text = strings.TrimSpace(text)
	var result strings.Builder
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' || r == '\'' {
			result.WriteRune(r)
		}
	}
	words := strings.Fields(result.String())
	return strings.Join(words, " ")
}

func flagOrEnv(name, value string) string {
	for i, a := range os.Args {
		if a == "-"+name && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return value
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
