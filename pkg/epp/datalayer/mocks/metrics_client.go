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

package mocks

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/gateway-api-inference-extension/pkg/epp/datalayer"
	logutil "sigs.k8s.io/gateway-api-inference-extension/pkg/epp/util/logging"
)

// MetricsClient provides a mock implementation of a metric fetching client.
type MetricsClient struct {
	mu        sync.RWMutex
	errmap    map[string]error
	resultmap map[string]*datalayer.Metrics
}

// NewMetricsClient returns a new mock metrics client initialized with the given
// results and/or errors.
func NewMetricsClient(results map[types.NamespacedName]*datalayer.Metrics, errmap map[types.NamespacedName]error) *MetricsClient {
	client := &MetricsClient{}
	client.SetResults(results)
	client.SetErrors(errmap)
	return client
}

// FetchMetrics returns the metrics (or error) associated with eth endpoint by looking
// up its keys in the results and errors maps.
func (cl *MetricsClient) FetchMetrics(ctx context.Context, ep datalayer.Addressable, last *datalayer.Metrics, _ int32) (*datalayer.Metrics, error) {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	key := ep.GetNamespacedName().String() // @TODO GetIPAddress() instead?
	if err, found := cl.errmap[key]; found {
		return nil, err
	}

	result, found := cl.resultmap[key]
	if !found {
		return nil, fmt.Errorf("no pod found: %v", ep.GetNamespacedName())
	}
	log.FromContext(ctx).V(logutil.VERBOSE).Info("Fetching metrics for pod",
		"name", ep.GetNamespacedName(), "existing", last, "new", result)
	return result.Clone(), nil
}

// SetResults sets the available results per endpoint.
func (cl *MetricsClient) SetResults(resultmap map[types.NamespacedName]*datalayer.Metrics) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.resultmap = make(map[string]*datalayer.Metrics, len(resultmap))

	for nsn, metric := range resultmap {
		cl.resultmap[nsn.String()] = metric
	}
}

// SetErrors sets the available errors per endpoint.
func (cl *MetricsClient) SetErrors(errmap map[types.NamespacedName]error) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.errmap = make(map[string]error, len(errmap))
	for nsn, err := range errmap {
		cl.errmap[nsn.String()] = err
	}
}
