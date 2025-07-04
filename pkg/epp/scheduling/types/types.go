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

package types

import (
	"fmt"

	backendmetrics "sigs.k8s.io/gateway-api-inference-extension/pkg/epp/backend/metrics"
	dltypes "sigs.k8s.io/gateway-api-inference-extension/pkg/epp/datalayer/types"
)

// LLMRequest is a structured representation of the fields we parse out of the LLMRequest body.
type LLMRequest struct {
	// RequestId is the Envoy generated Id for the request being processed
	RequestId string
	// TargetModel is the final target model after traffic split.
	TargetModel string
	// Prompt is the prompt that was sent in the request body.
	Prompt string
	// Headers is a map of the request headers.
	Headers map[string]string
}

func (r *LLMRequest) String() string {
	return fmt.Sprintf("RequestID: %s, TargetModel: %s, PromptLength: %d, Headers: %v", r.RequestId, r.TargetModel, len(r.Prompt), r.Headers)
}

type Pod interface {
	GetPod() *dltypes.PodInfo
	GetMetrics() *dltypes.Metrics
	String() string
}

type ScoredPod struct {
	Pod
	Score float64
}

func (pm *PodMetrics) String() string {
	if pm == nil {
		return ""
	}
	return fmt.Sprintf("%+v", *pm)
}

func (pm *PodMetrics) GetPod() *dltypes.PodInfo {
	return pm.PodInfo
}

func (pm *PodMetrics) GetMetrics() *dltypes.Metrics {
	return pm.Metrics
}

type PodMetrics struct {
	*dltypes.PodInfo
	*dltypes.Metrics
}

func ToSchedulerPodMetrics(pods []backendmetrics.PodMetrics) []Pod {
	pm := make([]Pod, 0, len(pods))
	for _, pod := range pods {
		pm = append(pm, &PodMetrics{PodInfo: pod.GetPod().Clone(), Metrics: pod.GetMetrics().Clone()})
	}
	return pm
}

// ProfileRunResult captures the profile run result.
type ProfileRunResult struct {
	TargetPod Pod
}

// SchedulingResult captures the result of the scheduling cycle.
type SchedulingResult struct {
	ProfileResults     map[string]*ProfileRunResult
	PrimaryProfileName string
}
