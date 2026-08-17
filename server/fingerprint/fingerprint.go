package fingerprint

import (
	"embed"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed rules.yaml
var rulesFS embed.FS

type FingerprintResult struct {
	IP         string  `json:"ip"`
	Port       int     `json:"port"`
	Protocol   string  `json:"protocol"`
	Product    string  `json:"product"`
	Version    string  `json:"version"`
	OsHint     string  `json:"os_hint"`
	Confidence float64 `json:"confidence"`
}

type BannerInput struct {
	IP     string `json:"ip"`
	Port   int    `json:"port"`
	Banner string `json:"banner"`
}

type Pattern struct {
	Regex         string  `yaml:"regex"`
	Product       string  `yaml:"product"`
	VersionGroup  int     `yaml:"version_group"`
	OsHintGroup   int     `yaml:"os_hint_group"`
	Confidence    float64 `yaml:"confidence"`
}

type OsHint struct {
	Match string `yaml:"match"`
	Hint  string `yaml:"hint"`
}

type Rule struct {
	Name     string    `yaml:"name"`
	Protocol string    `yaml:"protocol"`
	Patterns []Pattern `yaml:"patterns"`
	OsHints  []OsHint  `yaml:"os_hints"`
}

type RulesConfig struct {
	Rules []Rule `yaml:"rules"`
}

type Engine struct {
	config *RulesConfig
}

func NewEngine() (*Engine, error) {
	data, err := rulesFS.ReadFile("rules.yaml")
	if err != nil {
		return nil, err
	}

	var config RulesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &Engine{config: &config}, nil
}

func (e *Engine) Identify(input BannerInput) FingerprintResult {
	result := FingerprintResult{
		IP:         input.IP,
		Port:       input.Port,
		Protocol:   "unknown",
		Product:    "",
		Version:    "",
		OsHint:     "",
		Confidence: 0,
	}

	banner := input.Banner

	for _, rule := range e.config.Rules {
		for _, pattern := range rule.Patterns {
			re, err := regexp.Compile(pattern.Regex)
			if err != nil {
				continue
			}

			matches := re.FindStringSubmatch(banner)
			if matches != nil {
				if pattern.Confidence > result.Confidence {
					result.Protocol = rule.Protocol
					result.Product = pattern.Product
					result.Confidence = pattern.Confidence

					if pattern.VersionGroup > 0 && pattern.VersionGroup < len(matches) {
						result.Version = matches[pattern.VersionGroup]
					}

					if pattern.OsHintGroup > 0 && pattern.OsHintGroup < len(matches) {
						osRaw := strings.TrimSpace(matches[pattern.OsHintGroup])
						result.OsHint = e.extractOsHint(osRaw, rule.OsHints)
					}
				}
			}
		}
	}

	if result.OsHint == "" && result.Protocol != "unknown" {
		result.OsHint = e.extractOsHintFromBanner(banner, e.getAllOsHints())
	}

	return result
}

func (e *Engine) extractOsHint(text string, hints []OsHint) string {
	for _, hint := range hints {
		if strings.Contains(text, hint.Match) {
			return hint.Hint
		}
	}
	return text
}

func (e *Engine) extractOsHintFromBanner(banner string, hints []OsHint) string {
	for _, hint := range hints {
		if strings.Contains(banner, hint.Match) {
			return hint.Hint
		}
	}
	return ""
}

func (e *Engine) getAllOsHints() []OsHint {
	var allHints []OsHint
	for _, rule := range e.config.Rules {
		allHints = append(allHints, rule.OsHints...)
	}
	return allHints
}

func (e *Engine) IdentifyBatch(inputs []BannerInput) []FingerprintResult {
	results := make([]FingerprintResult, len(inputs))
	for i, input := range inputs {
		results[i] = e.Identify(input)
	}
	return results
}
