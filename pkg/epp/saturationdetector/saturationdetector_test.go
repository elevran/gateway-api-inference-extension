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

package saturationdetector

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"

	backendmetrics "sigs.k8s.io/gateway-api-inference-extension/pkg/epp/backend/metrics"
	"sigs.k8s.io/gateway-api-inference-extension/pkg/epp/datalayer/mocks"
	dltypes "sigs.k8s.io/gateway-api-inference-extension/pkg/epp/datalayer/types"
)

// --- Mock Implementations ---
type mockDatastore struct {
	pods []*mocks.Endpoint
}

// PodGetAll returns all pod metrics from the fake datastore.
// @TODO change to return datalayer.Endpoint:
// This is an adapter since Datastore expects backendmetrics.PodMetrics
// and we are moving to datalayer types.
func (mds *mockDatastore) PodGetAll() []backendmetrics.PodMetrics {
	result := make([]backendmetrics.PodMetrics, len(mds.pods))
	for i, ep := range mds.pods {
		result[i] = ep
	}
	return result
}

func newMockEndpoint(name string, metrics *dltypes.Metrics) *mocks.Endpoint {
	pod := &dltypes.PodInfo{
		NamespacedName: types.NamespacedName{Name: name, Namespace: "ns1"},
	}
	return mocks.NewEndpoint(pod, metrics)
}

// --- Tests ---

func TestNewDetector(t *testing.T) {
	tests := []struct {
		name                         string
		config                       *Config
		datastore                    Datastore
		expectedQueueDepthThreshold  int
		expectedKVCacheUtilThreshold float64
		expectedStalenessThreshold   time.Duration
	}{
		{
			name: "Valid config",
			config: &Config{
				QueueDepthThreshold:       10,
				KVCacheUtilThreshold:      0.8,
				MetricsStalenessThreshold: 100 * time.Millisecond,
			},
			datastore:                    &mockDatastore{},
			expectedQueueDepthThreshold:  10,
			expectedKVCacheUtilThreshold: 0.8,
			expectedStalenessThreshold:   100 * time.Millisecond,
		},
		{
			name: "invalid thresholds, fallback to default",
			config: &Config{
				QueueDepthThreshold:       -1,
				KVCacheUtilThreshold:      -5,
				MetricsStalenessThreshold: 0,
			},
			datastore:                    &mockDatastore{},
			expectedQueueDepthThreshold:  DefaultQueueDepthThreshold,
			expectedKVCacheUtilThreshold: DefaultKVCacheUtilThreshold,
			expectedStalenessThreshold:   DefaultMetricsStalenessThreshold,
		},
		{
			name: "kv cache threshold above range, fallback to default",
			config: &Config{
				QueueDepthThreshold:       10,
				KVCacheUtilThreshold:      1.5,
				MetricsStalenessThreshold: 100 * time.Millisecond,
			},
			datastore:                    &mockDatastore{},
			expectedQueueDepthThreshold:  10,
			expectedKVCacheUtilThreshold: DefaultKVCacheUtilThreshold,
			expectedStalenessThreshold:   100 * time.Millisecond,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// validate configuration values are loaded from env vars properly, including the use of default values when provided value is invalid.
			os.Setenv(EnvSdQueueDepthThreshold, strconv.Itoa(test.config.QueueDepthThreshold))
			os.Setenv(EnvSdKVCacheUtilThreshold, fmt.Sprintf("%v", test.config.KVCacheUtilThreshold))
			os.Setenv(EnvSdMetricsStalenessThreshold, test.config.MetricsStalenessThreshold.String())

			detector := NewDetector(LoadConfigFromEnv(), test.datastore, logr.Discard())
			if detector == nil {
				t.Fatalf("NewDetector() returned nil detector for valid config")
			}
			if detector.config.QueueDepthThreshold != test.expectedQueueDepthThreshold {
				t.Errorf("NewDetector() QueueDepthThreshold = %d, want %d", detector.config.QueueDepthThreshold, test.expectedQueueDepthThreshold)
			}
			if detector.config.KVCacheUtilThreshold != test.expectedKVCacheUtilThreshold {
				t.Errorf("NewDetector() KVCacheUtilThreshold = %f, want %f", detector.config.KVCacheUtilThreshold, test.expectedKVCacheUtilThreshold)
			}
			if detector.config.MetricsStalenessThreshold != test.expectedStalenessThreshold {
				t.Errorf("NewDetector() MetricsStalenessThreshold = %v, want %v", detector.config.MetricsStalenessThreshold, test.expectedStalenessThreshold)
			}
		})
	}
}

func TestDetector_IsSaturated(t *testing.T) {
	baseTime := time.Now()
	defaultConfig := &Config{
		QueueDepthThreshold:       5,
		KVCacheUtilThreshold:      0.90,
		MetricsStalenessThreshold: 100 * time.Millisecond,
	}

	tests := []struct {
		name            string
		config          *Config
		pods            []*mocks.Endpoint
		expectSaturated bool
	}{
		{
			name:            "No pods in datastore",
			config:          defaultConfig,
			pods:            []*mocks.Endpoint{},
			expectSaturated: true, // No capacity = saturated
		},
		{
			name:   "Single pod with good capacity",
			config: defaultConfig,
			pods: []*mocks.Endpoint{
				newMockEndpoint("pod1", &dltypes.Metrics{
					UpdateTime:          baseTime,
					WaitingQueueSize:    2,
					KVCacheUsagePercent: 0.5,
				}),
			},
			expectSaturated: false,
		},
		{
			name:   "Single pod with stale metrics",
			config: defaultConfig,
			pods: []*mocks.Endpoint{
				newMockEndpoint("pod1", &dltypes.Metrics{
					UpdateTime:          baseTime.Add(-200 * time.Millisecond), // Stale
					WaitingQueueSize:    1,
					KVCacheUsagePercent: 0.1,
				}),
			},
			expectSaturated: true,
		},
		{
			name:   "Single pod with high queue depth",
			config: defaultConfig,
			pods: []*mocks.Endpoint{
				newMockEndpoint("pod1", &dltypes.Metrics{
					UpdateTime:          baseTime,
					WaitingQueueSize:    10, // Exceeds threshold 5
					KVCacheUsagePercent: 0.1,
				}),
			},
			expectSaturated: true,
		},
		{
			name:   "Single pod with high KV cache utilization",
			config: defaultConfig,
			pods: []*mocks.Endpoint{
				newMockEndpoint("pod1", &dltypes.Metrics{
					UpdateTime:          baseTime,
					WaitingQueueSize:    1,
					KVCacheUsagePercent: 0.95, // Exceeds threshold 0.90
				}),
			},
			expectSaturated: true,
		},
		{
			name:   "Single pod with nil metrics",
			config: defaultConfig,
			pods: []*mocks.Endpoint{
				newMockEndpoint("pod1", nil),
			},
			expectSaturated: true,
		},
		{
			name:   "Multiple pods, all good capacity",
			config: defaultConfig,
			pods: []*mocks.Endpoint{
				newMockEndpoint("pod1", &dltypes.Metrics{
					UpdateTime:          baseTime,
					WaitingQueueSize:    1,
					KVCacheUsagePercent: 0.1,
				}),
				newMockEndpoint("pod2", &dltypes.Metrics{
					UpdateTime:          baseTime.Add(-10 * time.Millisecond),
					WaitingQueueSize:    0,
					KVCacheUsagePercent: 0.2,
				}),
			},
			expectSaturated: false,
		},
		{
			name:   "Multiple pods, one good, one bad (stale)",
			config: defaultConfig,
			pods: []*mocks.Endpoint{
				newMockEndpoint("pod1", &dltypes.Metrics{
					UpdateTime:          baseTime, // Good
					WaitingQueueSize:    1,
					KVCacheUsagePercent: 0.1,
				}),
				newMockEndpoint("pod2", &dltypes.Metrics{
					UpdateTime:          baseTime.Add(-300 * time.Millisecond), // Stale
					WaitingQueueSize:    0,
					KVCacheUsagePercent: 0.2,
				}),
			},
			expectSaturated: false, // One good pod is enough
		},
		{
			name:   "Multiple pods, one good, one bad (high queue)",
			config: defaultConfig,
			pods: []*mocks.Endpoint{
				newMockEndpoint("pod1", &dltypes.Metrics{
					UpdateTime:          baseTime,
					WaitingQueueSize:    1,
					KVCacheUsagePercent: 0.1,
				}),
				newMockEndpoint("pod2", &dltypes.Metrics{
					UpdateTime:          baseTime,
					WaitingQueueSize:    15, // Bad queue
					KVCacheUsagePercent: 0.2,
				}),
			},
			expectSaturated: false,
		},
		{
			name:   "Multiple pods, all bad capacity",
			config: defaultConfig,
			pods: []*mocks.Endpoint{
				newMockEndpoint("pod1", &dltypes.Metrics{
					UpdateTime:          baseTime.Add(-200 * time.Millisecond), // Stale
					WaitingQueueSize:    1,
					KVCacheUsagePercent: 0.1,
				}),
				newMockEndpoint("pod2", &dltypes.Metrics{
					UpdateTime:          baseTime,
					WaitingQueueSize:    20, // High queue
					KVCacheUsagePercent: 0.2,
				}),
				newMockEndpoint("pod3", &dltypes.Metrics{
					UpdateTime:          baseTime,
					WaitingQueueSize:    1,
					KVCacheUsagePercent: 0.99, // High KV
				}),
			},
			expectSaturated: true,
		},
		{
			name:   "Queue depth exactly at threshold",
			config: defaultConfig,
			pods: []*mocks.Endpoint{
				newMockEndpoint("pod1", &dltypes.Metrics{
					UpdateTime:          baseTime,
					WaitingQueueSize:    defaultConfig.QueueDepthThreshold, // Exactly at threshold (good)
					KVCacheUsagePercent: 0.1,
				}),
			},
			expectSaturated: false,
		},
		{
			name:   "KV cache exactly at threshold",
			config: defaultConfig,
			pods: []*mocks.Endpoint{
				newMockEndpoint("pod1", &dltypes.Metrics{
					UpdateTime:          baseTime,
					WaitingQueueSize:    1,
					KVCacheUsagePercent: defaultConfig.KVCacheUtilThreshold, // Exactly at threshold (good)
				}),
			},
			expectSaturated: false,
		},
		{
			name:   "Metrics age just over staleness threshold",
			config: defaultConfig,
			pods: []*mocks.Endpoint{
				newMockEndpoint("pod1", &dltypes.Metrics{
					UpdateTime:          baseTime.Add(-defaultConfig.MetricsStalenessThreshold - time.Nanosecond), // Just over (stale)
					WaitingQueueSize:    1,
					KVCacheUsagePercent: 0.1,
				}),
			},
			expectSaturated: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			detector := NewDetector(test.config, &mockDatastore{pods: test.pods}, logr.Discard())

			if got := detector.IsSaturated(context.Background()); got != test.expectSaturated {
				t.Errorf("IsSaturated() = %v, want %v", got, test.expectSaturated)
			}
		})
	}
}
