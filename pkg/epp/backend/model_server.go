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
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/gateway-api-inference-extension/pkg/epp/backend/k8s"
	"sigs.k8s.io/gateway-api-inference-extension/pkg/epp/backend/metrics"
)

// ModelServer represents attributes associated with a model server endpoint.
// This includes well known attributes such as Pod information and latest
// metrics. In addition, data collectors can attach arbitrary additional
// metadata or traits across different aspects of a ModelServer.
type ModelServer interface {
	fmt.Stringer
	// Pod handling
	GetPod() *k8s.PodInfo
	UpdatePod(*corev1.Pod)
	// metrics handling
	GetMetrics() *metrics.MetricsState
	UpdateMetrics(*metrics.MetricsState)
	// extended attributes
	GetAttribute(name string) (Cloneable, bool)
	UpdateAttribute(name string, value Cloneable)
}

// model server implements the ModelServer interface.
// It is intended as a data structure (e.g., passive, only holds state).
// Access is Go routine safe: either by use of atomic.Pointer or a safe
// map (access protected internally by mutex). Different concurrency schemes
// could be considered after benchmarking (e.g., sync.Map, moving the mutex
// to model server to protect all access, etc.).
type modelServer struct {
	pm    metrics.PodMetrics
	xattr attributes
}

// NewModelServer()
func NewModelServer() ModelServer {
	return &modelServer{
		xattr: *newAttributes(),
	}
}

// String returns a string representation of the model server attributes.
// For brevity, only the names of the extended attributes are returned,
// without their values.
func (ms *modelServer) String() string {
	return fmt.Sprintf("Pod: %v; Metrics: %v; Attr: %v", ms.pm.GetPod(), ms.pm.GetMetrics(), ms.xattr.keys())
}

// GetPod returns the Pod attributes associated with the model server.
func (ms *modelServer) GetPod() *k8s.PodInfo {
	return ms.pm.GetPod()
}

// UpdatePod updates the Pod attributes associated with the model server.
func (ms *modelServer) UpdatePod(pod *corev1.Pod) {
	ms.pm.UpdatePod(pod)
}

// GetMetrics returns the metrics associated with the model server.
func (ms *modelServer) GetMetrics() *metrics.MetricsState {
	return ms.pm.GetMetrics()
}

// UpdateMetrics updates the metrics associated with the model server.
func (ms *modelServer) UpdateMetrics(_ *metrics.MetricsState) {
	// TODO: implement. Currently handed internally by PodMetrics so unused by callers.
}

// GetAttribute returns the named extended attributes of the model server, if found.
func (ms *modelServer) GetAttribute(name string) (Cloneable, bool) {
	return ms.xattr.get(name)
}

// UpdateAttribute adds the named attribute to the available attributes of this model server.
func (ms *modelServer) UpdateAttribute(name string, value Cloneable) {
	ms.xattr.put(name, value)
}

// StopRefreshLoop is a temporary workaround to allow modelServer to
// be used as metrics.PodMetrics interface.
func (ms *modelServer) StopRefreshLoop() {
	ms.pm.StopRefreshLoop()
}
