// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package apphostingschema provides functionality around parsing and managing apphosting.yaml.
package apphostingschema

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

var (
	validAvailabilityValues = map[string]bool{"BUILD": true, "RUNTIME": true}
)

// AppHostingSchema is the struct representation of apphosting.yaml.
type AppHostingSchema struct {
	RunConfig RunConfig             `yaml:"runConfig,omitempty"`
	Env       []EnvironmentVariable `yaml:"env,omitempty"`
}

// RunConfig is the struct representation of the passed run config.
type RunConfig struct {
	// value types used must match server field types. pointers are used to capture unset vs zero-like values.
	CPU          *float32 `yaml:"cpu"`
	MemoryMiB    *int32   `yaml:"memoryMiB"`
	Concurrency  *int32   `yaml:"concurrency"`
	MaxInstances *int32   `yaml:"maxInstances"`
	MinInstances *int32   `yaml:"minInstances"`
}

// EnvironmentVariable is the struct representation of the passed environment variables.
type EnvironmentVariable struct {
	Variable     string   `yaml:"variable"`
	Value        string   `yaml:"value,omitempty"`  // Optional: Can be value xor secret
	Secret       string   `yaml:"secret,omitempty"` // Optional: Can be value xor secret
	Availability []string `yaml:"availability,omitempty"`
}

// UnmarshalYAML provides custom validation logic to validate EnvironmentVariable
func (ev *EnvironmentVariable) UnmarshalYAML(unmarshal func(any) error) error {
	type plain EnvironmentVariable // Define an alias
	if err := unmarshal((*plain)(ev)); err != nil {
		return err
	}

	if ev.Value != "" && ev.Secret != "" {
		return fmt.Errorf("both 'value' and 'secret' fields cannot be present")
	}

	if ev.Value == "" && ev.Secret == "" {
		return fmt.Errorf("either 'value' or 'secret' field is required")
	}

	for _, val := range ev.Availability {
		if !validAvailabilityValues[val] {
			return fmt.Errorf("invalid value in 'availability': %s", val)
		}
	}

	return nil
}

// UnmarshalYAML provides custom validation logic to validate RunConfig
func (rc *RunConfig) UnmarshalYAML(unmarshal func(any) error) error {
	type plain RunConfig // Define an alias
	if err := unmarshal((*plain)(rc)); err != nil {
		return err
	}

	// Validation for 'CPU'
	if rc.CPU != nil && !(1 <= *rc.CPU && *rc.CPU <= 8) {
		return fmt.Errorf("runConfig.cpu field is not in valid range of [1, 8]")
	}

	// Validation for 'MemoryMiB'
	if rc.MemoryMiB != nil && !(512 <= *rc.MemoryMiB && *rc.MemoryMiB <= 32768) {
		return fmt.Errorf("runConfig.memory field is not in valid range of [512, 32768]")
	}

	// Validation for 'Concurrency'
	if rc.Concurrency != nil && !(1 <= *rc.Concurrency && *rc.Concurrency <= 1000) {
		return fmt.Errorf("runConfig.concurrency field is not in valid range of [1, 1000]")
	}

	// Validation for 'MaxInstances'
	if rc.MaxInstances != nil && !(1 <= *rc.MaxInstances && *rc.MaxInstances <= 100) {
		return fmt.Errorf("runConfig.maxInstances field is not in valid range of [1, 100]")
	}

	// Validation for 'minInstances'
	if rc.MinInstances != nil && !(0 <= *rc.MinInstances && *rc.MinInstances <= 100) {
		return fmt.Errorf("runConfig.minInstances field is not in valid range of [1, 100]")
	}

	return nil
}

// ReadAndValidateAppHostingSchemaFromFile converts the provided file into an AppHostingSchema.
// Returns an empty AppHostingSchema{} if the file does not exist.
func ReadAndValidateAppHostingSchemaFromFile(filePath string) (AppHostingSchema, error) {
	a := AppHostingSchema{}
	apphostingBuffer, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		log.Printf("Missing apphosting config at %v, using reasonable defaults\n", filePath)
		return a, nil
	} else if err != nil {
		return a, fmt.Errorf("reading apphosting config at %v: %w", filePath, err)
	}

	if err = yaml.Unmarshal(apphostingBuffer, &a); err != nil {
		return a, fmt.Errorf("unmarshalling apphosting config as YAML: %w", err)
	}
	return a, nil
}
