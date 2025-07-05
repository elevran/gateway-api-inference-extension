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
	"fmt"
	"sync/atomic"

	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/gateway-api-inference-extension/pkg/epp/datalayer"
)

type Endpoint struct {
	Pod     atomic.Pointer[datalayer.PodInfo]
	Metrics atomic.Pointer[datalayer.Metrics]
	attr    datalayer.Attributes
}

// NewEndpoint returns a new mock Endpoint instance initialed with the
// provided PodInfo and Metrics.
func NewEndpoint(pod *datalayer.PodInfo, metrics *datalayer.Metrics) *Endpoint {
	ep := &Endpoint{
		attr: *datalayer.NewAttributes(),
	}
	ep.Pod.Store(pod) // optionally create a corev1.Pod object and call UpdatePod?
	ep.Metrics.Store(metrics)
	return ep
}

// String returns a string representation of the mock Endpoint.
func (ep *Endpoint) String() string {
	return fmt.Sprintf("Pod: %v; Metrics: %v; Attributes: %v", ep.GetPod(), ep.GetMetrics(), ep.Keys())
}

// GetPod returns the PodInfo of the endpoint.
func (ep *Endpoint) GetPod() *datalayer.PodInfo {
	return ep.Pod.Load()
}

// UpdatePod sets the PodInfo field of the endpoint
func (ep *Endpoint) UpdatePod(pod *corev1.Pod) {
	ep.Pod.Store(datalayer.FromAPIPod(pod))
}

// GetMetrics returns the Metrics of the endpoint.
func (ep *Endpoint) GetMetrics() *datalayer.Metrics {
	return ep.Metrics.Load()
}

// UpdateMetrics sets the Metrics field of the endpoint.
func (ep *Endpoint) UpdateMetrics(metrics *datalayer.Metrics) {
	ep.Metrics.Store(metrics)
}

// Get returns an attribute value for the named key.
func (ep *Endpoint) Get(key string) (datalayer.Cloneable, bool) {
	return ep.attr.Get(key)
}

// Keys returns an array of attribute names associated with the endpoint.
func (ep *Endpoint) Keys() []string {
	return ep.attr.Keys()
}

// Put sets an attribute value associated with the endpoint.
func (ep *Endpoint) Put(key string, val datalayer.Cloneable) {
	ep.attr.Put(key, val)
}

// StopRefreshLoop is a no-op of mock Endpoint instances.
func (ep *Endpoint) StopRefreshLoop() {
	// noop
}
