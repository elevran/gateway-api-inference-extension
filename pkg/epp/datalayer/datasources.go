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

package datalayer

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	dltypes "sigs.k8s.io/gateway-api-inference-extension/pkg/epp/datalayer/types"
)

var (
	// DefaultDataSources is the system default data source registry.
	DefaultDataSources = DataSourceRegistry{}
)

// DataSourceRegistry stores named data sources and makes them
// accessible to GIE subsystems
type DataSourceRegistry struct {
	mu      sync.RWMutex
	sources map[string]dltypes.DataSource
}

// RegisterSource adds a source to the registry.
func (dsr *DataSourceRegistry) RegisterSource(src dltypes.DataSource) error {
	if src == nil {
		return errors.New("unable to register a nil data source")
	}

	dsr.mu.Lock()
	defer dsr.mu.Unlock()

	if _, found := dsr.sources[src.Name()]; found {
		return fmt.Errorf("unable to register duplicate data source: %s", src.Name())
	}
	dsr.sources[src.Name()] = src
	return nil
}

// GetNamedDataSource returns the named data source, if found.
func (dsr *DataSourceRegistry) GetNamedDataSource(name string) (dltypes.DataSource, error) {
	if name == "" {
		return nil, errors.New("unable to retrieve a data source without a name")
	}

	dsr.mu.RLock()
	defer dsr.mu.RUnlock()
	if ds, found := dsr.sources[name]; found {
		return ds, nil
	}
	return nil, &NotFoundError{Name: name}
}

// RegisterSource adds the data source to the default registry.
func RegisterSource(src dltypes.DataSource) error {
	return DefaultDataSources.RegisterSource(src)
}

// GetNamedDataSource returns the named source from the
// default registry, if found.
func GetNamedDataSource(name string) (dltypes.DataSource, error) {
	return DefaultDataSources.GetNamedDataSource(name)
}

// NotFoundError is an explicit error value raised when a
// source is not found in the requested registry.
type NotFoundError struct {
	Name string
}

// Error returns tha associated error string.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("data source not found: %v", e.Name)
}

// ValidateExtractorType checks if an extractor can handle
// the collector's output.
func ValidateExtractorType(collectorOutputType, extractorInputType reflect.Type) error {
	if collectorOutputType == extractorInputType {
		return nil
	}

	// extractor accepts anything (i.e., interface{})
	if extractorInputType.Kind() == reflect.Interface && extractorInputType.NumMethod() == 0 {
		return nil
	}

	// Check if collector output implements extractor input interface
	if collectorOutputType.Implements(extractorInputType) {
		return nil
	}

	return fmt.Errorf("extractor input type %v cannot handle collector output type %v",
		extractorInputType, collectorOutputType)
}
