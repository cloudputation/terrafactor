package terrafactor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ProviderState is the persistent snapshot written after every successful render.
type ProviderState struct {
	Provider   string              `json:"provider"`
	RenderedAt string              `json:"rendered_at"`
	Resources  map[string][]string `json:"resources"`
}

// LoadProviderState reads and decodes the state file. Returns an error if the file does not exist.
func LoadProviderState(stateFile string) (*ProviderState, error) {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, err
	}
	var state ProviderState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("invalid state file %s: %w", stateFile, err)
	}
	return &state, nil
}

// SaveProviderState encodes and writes the state file, creating parent directories as needed.
func SaveProviderState(stateFile string, state ProviderState) error {
	state.RenderedAt = time.Now().UTC().Format(time.RFC3339)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal provider state: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return err
	}
	return os.WriteFile(stateFile, data, 0644)
}

// resourceSpecToStateMap converts []ResourceSpec into the map[name]→[]fieldName form stored in state.
func resourceSpecToStateMap(resources []ResourceSpec) map[string][]string {
	m := make(map[string][]string, len(resources))
	for _, r := range resources {
		fields := make([]string, len(r.Fields))
		for i, f := range r.Fields {
			fields[i] = f.Name
		}
		sort.Strings(fields)
		m[r.ResourceName] = fields
	}
	return m
}

// diffHasFieldChanges returns true if the prev field name list differs from the next ResourceField list.
func diffHasFieldChanges(prev []string, next []ResourceField) bool {
	prevSet := make(map[string]struct{}, len(prev))
	for _, f := range prev {
		prevSet[f] = struct{}{}
	}
	nextSet := make(map[string]struct{}, len(next))
	for _, f := range next {
		nextSet[f.Name] = struct{}{}
	}
	if len(prevSet) != len(nextSet) {
		return true
	}
	for n := range nextSet {
		if _, ok := prevSet[n]; !ok {
			return true
		}
	}
	return false
}

// stateFilePath returns the canonical path to the provider state file.
func stateFilePath(rootDir, providerName string) string {
	return filepath.Join(rootDir, EngineBaseDir, providerName, ".terrafactor-state.json")
}
