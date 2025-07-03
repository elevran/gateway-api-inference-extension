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
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/gateway-api-inference-extension/pkg/epp/backend/metrics"
)

// modelServerFactory is a wrapper for extending the metrics.PodMetrics factory
// to return ModelServer.
type modelServerFactory struct {
	delegate metrics.PodMetricsFactory
}

// NewModelServerFactory returns a factory for ModelServes
func NewModelServerFactory(cl metrics.PodMetricsClient, interval time.Duration) metrics.PodMetricsFactory {
	return &modelServerFactory{
		delegate: metrics.NewPodMetricsFactory(cl, interval),
	}
}

// RefreshInterval returns the metric collection interval.
// @TODO: remove once decoupled DataSource is introduced.
func (msf *modelServerFactory) RefreshInterval() time.Duration {
	return msf.delegate.RefreshInterval()
}

// NewModelServer returns a new ModelServer for the specified Pod, adding it to the data store.
// @TODO temporary wrapper
func (msf *modelServerFactory) NewPodMetrics(parent context.Context, p *corev1.Pod, ds metrics.Datastore) metrics.PodMetrics {
	return msf.NewModelServer(parent, p, ds)
}

// NewModelServer returns a new model server
func (msf *modelServerFactory) NewModelServer(parent context.Context, p *corev1.Pod, ds Datastore) metrics.PodMetrics {
	return NewModelServer(msf.delegate.NewPodMetrics(parent, p, &datastoreAdapter{inner: ds}))
}
