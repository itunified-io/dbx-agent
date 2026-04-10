package template

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// MonitoringTemplate defines a set of checks and thresholds for a target type.
type MonitoringTemplate struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	TargetType  string            `yaml:"target_type"` // oracle, postgres, host
	Collectors  []CollectorSpec   `yaml:"collectors"`
	Thresholds  map[string]Threshold `yaml:"thresholds"`
}

// CollectorSpec defines a collector to enable and its interval.
type CollectorSpec struct {
	Name       string `yaml:"name"`
	Enabled    bool   `yaml:"enabled"`
	IntervalMs int    `yaml:"interval_ms"`
}

// Threshold defines warning/critical levels for a metric.
type Threshold struct {
	Warning  float64 `yaml:"warning"`
	Critical float64 `yaml:"critical"`
	Operator string  `yaml:"operator"` // gt, lt, eq
}

// Loader loads monitoring templates from YAML files.
type Loader struct {
	templates map[string]*MonitoringTemplate
}

// NewLoader creates a template loader.
func NewLoader() *Loader {
	return &Loader{templates: make(map[string]*MonitoringTemplate)}
}

// LoadDir loads all YAML templates from a directory.
func (l *Loader) LoadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read template dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if err := l.LoadFile(path); err != nil {
			return fmt.Errorf("load template %s: %w", path, err)
		}
	}

	return nil
}

// LoadFile loads a single YAML template file.
func (l *Loader) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var tmpl MonitoringTemplate
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	if tmpl.Name == "" {
		return fmt.Errorf("template missing name in %s", path)
	}
	l.templates[tmpl.Name] = &tmpl
	return nil
}

// Get returns a template by name.
func (l *Loader) Get(name string) (*MonitoringTemplate, bool) {
	t, ok := l.templates[name]
	return t, ok
}

// All returns all loaded templates.
func (l *Loader) All() map[string]*MonitoringTemplate {
	return l.templates
}

// MergeOverride applies per-target overrides on top of a base template.
func MergeOverride(base *MonitoringTemplate, overrides map[string]Threshold) *MonitoringTemplate {
	merged := *base
	merged.Thresholds = make(map[string]Threshold)
	for k, v := range base.Thresholds {
		merged.Thresholds[k] = v
	}
	for k, v := range overrides {
		merged.Thresholds[k] = v
	}
	return &merged
}
