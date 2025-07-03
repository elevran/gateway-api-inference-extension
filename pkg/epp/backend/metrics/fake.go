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
	"context"
	"fmt"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/gateway-api-inference-extension/pkg/epp/backend/k8s"
	logutil "sigs.k8s.io/gateway-api-inference-extension/pkg/epp/util/logging"
)

// FakePodMetrics is an implementation of PodMetrics that doesn't run the async refresh loop.
type FakePodMetrics struct {
	Pod     *k8s.PodInfo
	Metrics *MetricsState
}

func (fpm *FakePodMetrics) String() string {
	return fmt.Sprintf("Pod: %v; Metrics: %v", fpm.GetPod(), fpm.GetMetrics())
}

func (fpm *FakePodMetrics) GetPod() *k8s.PodInfo {
	return fpm.Pod
}
func (fpm *FakePodMetrics) GetMetrics() *MetricsState {
	return fpm.Metrics
}
func (fpm *FakePodMetrics) UpdatePod(pod *corev1.Pod) {
	fpm.Pod = k8s.FromAPIPod(pod)
}
func (fpm *FakePodMetrics) StopRefreshLoop() {} // noop

type FakePodMetricsClient struct {
	errMu sync.RWMutex
	// TODO: Fetching is done from an IP address so more likely that we should
	// be using string as the map key. Keep as is for compatibility with existing
	// tests which might not have defined and unique IP addresses.
	Err   map[types.NamespacedName]error
	resMu sync.RWMutex
	Res   map[types.NamespacedName]*MetricsState
}

func (f *FakePodMetricsClient) FetchMetrics(ctx context.Context, endpoint llmServer, existing *MetricsState, port int32) (*MetricsState, error) {
	id := endpoint.GetNamespacedName() // TODO: see above note on struct definition

	f.errMu.RLock()
	err, ok := f.Err[id]
	f.errMu.RUnlock()
	if ok {
		return nil, err
	}
	f.resMu.RLock()
	res, ok := f.Res[id]
	f.resMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no pod found: %v", id)
	}
	log.FromContext(ctx).V(logutil.VERBOSE).Info("Fetching metrics for pod", "existing", existing, "new", res)
	return res.Clone(), nil
}

func (f *FakePodMetricsClient) SetRes(new map[types.NamespacedName]*MetricsState) {
	f.resMu.Lock()
	defer f.resMu.Unlock()
	f.Res = new
}

func (f *FakePodMetricsClient) SetErr(new map[types.NamespacedName]error) {
	f.errMu.Lock()
	defer f.errMu.Unlock()
	f.Err = new
}
