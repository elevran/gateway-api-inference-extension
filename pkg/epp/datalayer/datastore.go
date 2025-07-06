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
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/gateway-api-inference-extension/api/v1alpha2"
)

// @TODO
// - Initial split of data store functionality by caller. This is subject to change!
// - Should all controller/reconcilers move to store as they're effectively tied to it?
//   This would help in making the data store interface more cohesive and reduce the API.

type InferencePoolSetter interface {
	// PoolSet sets the given pool in datastore. If the given pool has different label selector than the previous pool
	// that was stored, the function triggers a resync of the pods to keep the datastore updated. If the given pool
	// is nil, this call triggers the datastore.Clear() function.
	PoolSet(ctx context.Context, reader client.Reader, pool *v1alpha2.InferencePool) error
	// Clears the store state, happens when the pool gets deleted or reset.
	Clear()
}

type InferencePoolGetter interface {
	// InferencePool operations
	PoolGet() (*v1alpha2.InferencePool, error)
	PoolHasSynced() bool
	PoolLabelsMatch(podLabels map[string]string) bool
}

type InferenceModelAccess interface {
	// InferenceModel operations
	ModelSetIfOlder(infModel *v1alpha2.InferenceModel) bool
	ModelDelete(namespacedName types.NamespacedName) *v1alpha2.InferenceModel

	ModelGet(modelName string) *v1alpha2.InferenceModel
	ModelResync(ctx context.Context, reader client.Reader, modelName string) (bool, error)
	ModelGetAll() []*v1alpha2.InferenceModel
}

type PodAccess interface {
	// PodMetrics operations
	PodUpdateOrAddIfNotExist(pod *corev1.Pod) bool
	PodDelete(namespacedName types.NamespacedName)

	// PodGetAll returns all pods and metrics, including fresh and stale.
	PodGetAll() []Endpoint
	// PodList lists pods matching the given predicate.
	PodList(predicate func(Endpoint) bool) []Endpoint
}

type Datastore interface {
	InferencePoolSetter
	InferencePoolGetter
	InferenceModelAccess
	PodAccess
}
