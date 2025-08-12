package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ModelMap struct {
	Name string `json:"name"`
}
type Provider struct {
	DefaultModel string              `json:"defaultModel,omitempty"`
	Models       map[string]ModelMap `json:"models"`
}
type Config struct {
	Provider map[string]Provider `json:"provider"`
}

// Simple per-1M token pricing (USD)
var inPrice = map[string]float64{
	"gpt-5":      1.25,
	"gpt-5-mini": 0.25,
	"gpt-5-nano": 0.05,
}
var outPrice = map[string]float64{
	"gpt-5":      10.00,
	"gpt-5-mini": 2.00,
	"gpt-5-nano": 0.40,
}

func approxTokens(chars int) int {
	if chars <= 0 {
		return 1
	}
	// ~4 chars per token (very rough)
	t := chars / 4
	if t < 1 {
		return 1
	}
	return t
}
func estimateCostUSD(model string, inputChars, outputChars int) float64 {
	inT := float64(approxTokens(inputChars))
	outT := float64(approxTokens(outputChars))
	return (inT/1_000_000.0)*inPrice[model] + (outT/1_000_000.0)*outPrice[model]
}

func main() {
	model := flag.String("model", "gpt-5-mini", "gpt-5 | gpt-5-mini | gpt-5-nano")
	mode := flag.String("mode", "code", "plan | code | debug")
	filePaths := multiFlag{}
	flag.Var(&filePaths, "file", "file(s) to add as context (repeatable)")
	maxOut := flag.Int("max-tokens", 0, "cap output tokens")
	temperature := flag.Float64("temperature", 0.2, "sampling temperature")
	flag.Parse()

	prompt := strings.TrimSpace(strings.Join(flag.Args(), " "))
	if prompt == "" {
		fmt.Fprintln(os.Stderr, "usage: ocx --model gpt-5-mini --mode plan \"your prompt\"")
		os.Exit(2)
	}

	if os.Getenv("OPENAI_API_KEY") == "" {
		fmt.Fprintln(os.Stderr, "OPENAI_API_KEY not set")
		os.Exit(2)
	}

	cfg := loadPreset()
	if _, ok := cfg.Provider["openai"].Models[*model]; !ok {
		die(fmt.Errorf("unknown model: %s", *model))
	}
	cfg.Provider["openai"].DefaultModel = *model
	writeConfig(cfg)

	builder := strings.Builder{}
	builder.WriteString(prompt)

	if len(filePaths) > 0 {
		builder.WriteString("\n\nContext files:\n")
		for _, p := range filePaths {
			builder.WriteString("\n===== FILE: " + p + " =====\n")
			b, err := os.ReadFile(p)
			if err != nil {
				builder.WriteString("<<error reading " + err.Error() + ">>\n")
			} else {
				builder.Write(b)
				builder.WriteString("\n")
			}
		}
	}

	var system string
	switch *mode {
	case "plan":
		system = "You are a senior engineer. First output a concise numbered plan (trade-offs/risks), then the code."
	case "code":
		system = "You are a precise coding assistant. Output minimal text and runnable code."
	case "debug":
		system = "You are a debugger. Identify root cause, propose fix, then show a patch or replacement snippet."
	default:
		die(errors.New("mode must be plan|code|debug"))
	}

	// Pre-flight cost estimate (input only, rough)
	est := estimateCostUSD(*model, len(builder.String()), 0)
	fmt.Fprintf(os.Stderr, "[ocx] %s · %s · est $%.6f (input only)\n", *model, *mode, est)

	env := os.Environ()
	env = append(env, "OPENCODE_SYSTEM_HINT="+system)
	if *maxOut > 0 {
		env = append(env, fmt.Sprintf("OPENCODE_MAX_TOKENS=%d", *maxOut))
	}
	env = append(env, fmt.Sprintf("OPENCODE_TEMPERATURE=%.3f", *temperature))

	cmd := exec.Command("opencode")
	cmd.Env = env
	cmd.Stdin = strings.NewReader(builder.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		die(err)
	}
}

func loadPreset() Config {
	exe, _ := os.Executable()
	// repo root assumption: .../cmd/ocx/ocx
	root := filepath.Dir(filepath.Dir(filepath.Dir(exe)))
	path := filepath.Join(root, "presets", "opencode.json")
	b, err := os.ReadFile(path)
	if err != nil {
		die(fmt.Errorf("missing presets/opencode.json: %w", err))
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		die(err)
	}
	return cfg
}

func writeConfig(cfg Config) {
	_ = os.MkdirAll(".opencode", 0o755)
	out := filepath.Join(".opencode", "opencode.json")
	b, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(out, b, 0o644); err != nil {
		die(err)
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

type multiFlag []string
func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(val string) error { *m = append(*m, val); return nil }
