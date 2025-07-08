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
	"reflect"

	dltypes "sigs.k8s.io/gateway-api-inference-extension/pkg/epp/datalayer/types"
)

const (
	ExtractorName = "model-server-protocol-metrics"
)

// Extractor is used to extract the specific metrics defined
// in datalayer.Metrics from the collector's output (which holds
// the full set of metrics)
type Extractor struct {
	// TBD
}

// Name returns the name of this extractor
func (ext *Extractor) Name() string {
	return ExtractorName
}

// ExpectedType defines the type expected by the extractor. It must match
// the DataSource.OutputType() the extractor registers for.
func (ext *Extractor) ExpectedType() reflect.Type {
	return reflect.TypeOf(Collection{})
}

// Extract transforms the data source output into a concrete attribute that
// is stored on the given endpoint.
func (ext *Extractor) Extract(data any, ep dltypes.Endpoint) {
	// start with the current endpoint metrics since we prefer stale over none...
	// convert the data to a metrics.Collection and then
	// use metrics.Spec and metrics.State to extract the actual
	// metrics and store on the endpoint
}
