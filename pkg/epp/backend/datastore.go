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

package backend

import (
	"sigs.k8s.io/gateway-api-inference-extension/api/v1alpha2"
	"sigs.k8s.io/gateway-api-inference-extension/pkg/epp/backend/metrics"
)

// @TODO change to use ModelServer.
// The below is the interface expected by metrics.Datastore
type Datastore interface {
	PoolGet() (*v1alpha2.InferencePool, error)
	PodGetAll() []metrics.PodMetrics
	PodList(func(metrics.PodMetrics) bool) []metrics.PodMetrics
}

type datastoreAdapter struct {
	inner Datastore
}

func (dsa *datastoreAdapter) PoolGet() (*v1alpha2.InferencePool, error) {
	return dsa.inner.PoolGet()
}

func (dsa *datastoreAdapter) PodGetAll() []metrics.PodMetrics {
	return dsa.inner.PodGetAll()
}

func (dsa *datastoreAdapter) PodList(predicate func(metrics.PodMetrics) bool) []metrics.PodMetrics {
	return dsa.inner.PodList(predicate)
}

/*
func (d *datastoreAdapter) PodGetAll() []metrics.PodMetrics {
    orig := d.inner.PodGetAll()
    out := make([]metrics.PodMetrics, len(orig))
    for i, o := range orig {
        out[i] = o // ModelServer implements PodMetrics
    }
    return out
}

func (d *datastoreAdapter) PodList(pred func(metrics.PodMetrics) bool) []metrics.PodMetrics {
    // Adapt predicate to ModelServer
    wrappedPred := func(ms backend.ModelServer) bool {
        return pred(ms) // ModelServer implements PodMetrics
    }

    orig := d.inner.PodList(wrappedPred)
    out := make([]metrics.PodMetrics, len(orig))
    for i, o := range orig {
        out[i] = o
    }
    return out
}
*/
