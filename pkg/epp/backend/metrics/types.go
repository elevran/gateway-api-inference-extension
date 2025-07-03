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

// Package metrics is a library to interact with backend metrics.
package metrics

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api-inference-extension/pkg/epp/backend/k8s"
)

type PodMetrics interface {
	GetPod() *k8s.PodInfo
	GetMetrics() *MetricsState
	UpdatePod(*corev1.Pod)
	StopRefreshLoop()
	String() string
}

type PodMetricsFactory interface {
	RefreshInterval() time.Duration
	NewPodMetrics(parent context.Context, p *corev1.Pod, ds Datastore) PodMetrics
}

// llmServer is an internal interface defining for expectations of
// endpoints passed to PodMetricsClient
type llmServer interface {
	GetNamespacedName() types.NamespacedName
	GetIPAddress() string
}

type PodMetricsClient interface {
	FetchMetrics(ctx context.Context, endpoint llmServer, existing *MetricsState, port int32) (*MetricsState, error)
}
