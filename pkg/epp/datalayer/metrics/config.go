/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"encoding/json"
	"time"
)

const (
	// Per endpoint configuration:
	// Default metrics collection interval
	DefaultCollectionInterval = 50 * time.Millisecond
	// Default metrics validity period
	DefaultValidityPeriod = 5 * time.Second
	// Global data source debug configuration
	// Default global metric collection interval
	DefaultDataSourceMetricCollectionInterval = 5 * time.Second
	// Debug print interval (fixed)
	DataSourceDebugPrintInterval = 5 * time.Second
)

// DataSourceConfig holds the configuration for the metrics data source.
type DataSourceConfig struct {
	// Interval controls how often per-endpoint metrics are refreshed.
	Interval time.Duration `json:"interval"`
	// SourceMetrics controls if and how global data source metrics
	// are collected and published.
	SourceMetrics *DataSourceMetrics `json:"datasource"`
	// Debug control if global metrics data source debug is produced.
	Debug *DataSourceDebug `json:"debug"`
}

// DataSourceMetrics holds configuration for global metrics, across all
// endpoints and the data source itself (i.e., as opposed to per endpoint
// metrics produced by the source and associated extractors).
type DataSourceMetrics struct {
	// Enable producing global data source metrics.
	Enabled bool `json:"enabled"`
	// Validity defines the period in which metrics are considered fresh.
	Validity time.Duration `json:"validity"`
	// Interval specifies how often to produce the global data source metrics.
	Interval time.Duration `json:"interval"`
}

// DataSourceDebug holds debug configuration of the metrics data source.
type DataSourceDebug struct {
	// Enabled is set to true if debug logs are enabled.
	Enabled bool `json:"enabled"`
}

// NewDefaultDataSourceConfig returns a default data source configuration.
// Global and debug information are disabled by default and must be
// explicitly set by the caller before creating the data source.
// Data source configuration can not be changed once the source is created.
func NewDefaultDataSourceConfig() *DataSourceConfig {
	return &DataSourceConfig{
		Interval: DefaultDataSourceMetricCollectionInterval,
		SourceMetrics: &DataSourceMetrics{
			Interval: DefaultDataSourceMetricCollectionInterval,
			Validity: DefaultValidityPeriod,
		},
		Debug: &DataSourceDebug{},
	}
}

// ConfigFromRawJSON returns a configuration based on the default and
// any overrides set in the JSON byte array.
func ConfigFromRawJSON(msg json.RawMessage) (*DataSourceConfig, error) {
	cfg := NewDefaultDataSourceConfig()
	if err := json.Unmarshal(msg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
