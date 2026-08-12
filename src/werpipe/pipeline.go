package werpipe

import (
	"bytes"
	"fmt"
	"math"
	"os/exec"
	"strings"
)

type Pipeline struct {
	WhisperCLI string
	ModelDir   string
	Threads    int
}

func NewPipeline(whisperCLI, modelDir string, threads int) *Pipeline {
	return &Pipeline{
		WhisperCLI: whisperCLI,
		ModelDir:   modelDir,
		Threads:    threads,
	}
}

func (p *Pipeline) Run(level QuantLevel, modelName string, samples []Sample) ([]SampleResult, error) {
	modelPath := modelName
	if !strings.HasPrefix(modelPath, "/") && p.ModelDir != "" {
		modelPath = p.ModelDir + "/" + modelName
	}
	threads := p.Threads
	if threads < 1 {
		threads = 4
	}

	results := make([]SampleResult, len(samples))
	for i, s := range samples {
		out, err := p.transcribe(modelPath, s.AudioPath, threads)
		results[i] = SampleResult{
			SampleID: s.ID,
			WER:      math.NaN(),
		}
		if err != nil {
			results[i].Error = err
			continue
		}
		hyp := Normalize(out)
		results[i].Hypothesis = hyp
		results[i].WER = ComputeWER(s.Reference, hyp)
		results[i].Error = nil
	}
	return results, nil
}

func (p *Pipeline) transcribe(modelPath, audioPath string, threads int) (string, error) {
	cmd := exec.Command(p.WhisperCLI,
		"-m", modelPath,
		"-f", audioPath,
		"--no-timestamps",
		"-l", "en",
		"-t", fmt.Sprintf("%d", threads),
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("whisper-cli: %w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}
