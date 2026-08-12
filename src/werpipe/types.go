package werpipe

type QuantLevel string

const (
	F16  QuantLevel = "f16"
	Q8_0 QuantLevel = "q8_0"
	Q5_0 QuantLevel = "q5_0"
	Q4_0 QuantLevel = "q4_0"
)

type Sample struct {
	ID        string
	AudioPath string
	Reference string
}

type SampleResult struct {
	SampleID   string
	Hypothesis string
	WER        float64
	Error      error
}

type LevelResults struct {
	Level      QuantLevel
	Samples    []SampleResult
	MeanWER    float64
	MedianWER  float64
	StdDev     float64
	NumSuccess int
	NumError   int
}

type Comparison struct {
	Level         QuantLevel
	BaselineWER   float64
	QuantWER      float64
	RelChangePct  float64
	PValue        float64
	Significant   bool
	BootstrapCI95 [2]float64
}
