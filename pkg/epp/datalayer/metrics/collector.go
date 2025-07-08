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
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/go-logr/logr"
	prometheus "github.com/prometheus/client_model/go"

	dltypes "sigs.k8s.io/gateway-api-inference-extension/pkg/epp/datalayer/types"
)

const (
	// The metrics data source name.
	DatasourceName = "metrics-datasource"
)

// Collection is a parsed set of Prometheus metrics produced by the collector.
type Collection map[string]*prometheus.MetricFamily

// Collector represents a metrics data source.
type Collector struct {
	mu         sync.RWMutex
	endpoints  map[string]*endpointState    // keyed by server IP
	extractors map[string]dltypes.Extractor // keyed by extractor name

	config    DataSourceConfig
	datastore dltypes.InferencePoolGetter
	logger    logr.Logger

	// Global context for entire collector
	ctx    context.Context
	cancel context.CancelFunc

	// Lifecycle management
	startOnce sync.Once
	stopOnce  sync.Once
	running   bool
}

// endpointState tracks the state of metrics collection for a single endpoint
type endpointState struct {
	server dltypes.Endpoint

	// Per-endpoint context and cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Collection state
	ticker *time.Ticker
	// TODO: reuse an http client
	lastUpdate time.Time
	errors     int
	maxErrors  int // TODO: fixed constant? ensure over some time or consecutive?

	// Goroutine management
	done chan struct{}
	wg   sync.WaitGroup
}

func NewDataSource(ctx context.Context, store dltypes.InferencePoolGetter, logger logr.Logger) (*Collector, error) {
	return NewDataSourceWithConfig(ctx, store, logger, NewDefaultDataSourceConfig())
}

func NewDataSourceWithConfig(
	ctx context.Context, store dltypes.InferencePoolGetter,
	logger logr.Logger, cfg *DataSourceConfig) (*Collector, error) {

	if ctx == nil {
		ctx = context.Background()
	}
	if cfg == nil {
		cfg = NewDefaultDataSourceConfig()
	}

	ctx, cancel := context.WithCancel(ctx)
	return &Collector{
		endpoints: make(map[string]*endpointState),
		datastore: store,
		config:    *cfg,
		logger:    logger,
		ctx:       ctx,
		cancel:    cancel,
		running:   false,
	}, nil
}

// Name returns the collector's name.
func (c *Collector) Name() string {
	return DatasourceName
}

// Start begins the collection process.
func (emc *Collector) Start(ctx context.Context) error {
	var err error

	emc.startOnce.Do(func() {
		emc.mu.Lock()
		defer emc.mu.Unlock()

		if emc.running {
			err = errors.New("collector already running")
			return
		}

		emc.running = true
		emc.logger.Info("starting data source metrics collector")

		for _, state := range emc.endpoints { // Start existing endpoints, if any
			emc.startEndpointCollectionLocked(state)
		}
		emc.logger.Info("started data source metrics collector")

		// TODO start the global metrics publisher and debug logger
	})
	return err
}

// Stop stops the collection process.
func (emc *Collector) Stop() error {
	emc.stopOnce.Do(func() {
		emc.logger.Info("stopping data source metrics collector")

		emc.mu.Lock()
		defer emc.mu.Unlock()

		emc.running = false
		emc.cancel() // Cancel global context first

		for _, state := range emc.endpoints { // Stop all per endpoint go routines
			emc.stopEndpointCollectionLocked(state)
		}
		emc.logger.Info("data source metrics collector stopped")
	})
	return nil
}

// OutputType returns the type of information collected.
func (*Collector) OutputType() reflect.Type {
	return reflect.TypeOf(Collection{})
}

// AddExtractor adds an extractor to the data source.
// The extractor will be called whenever the data source might
// have some new raw information regarding an endpoint.
// The Extractor's expected input type is validated when it is
// registered
func (emc *Collector) AddExtractor(extractor dltypes.Extractor) error {
	name := extractor.Name()
	emc.mu.Lock()
	defer emc.mu.Unlock()

	if _, exists := emc.extractors[name]; exists {
		return fmt.Errorf("extractor named %s already registered", name)
	}
	emc.extractors[name] = extractor
	return nil
}

// AddEndpoint adds an endpoint to collect from.
func (emc *Collector) AddEndpoint(ep dltypes.Endpoint) error {
	emc.mu.Lock()
	defer emc.mu.Unlock()

	pod := ep.GetPod()
	if pod == nil {
		return errors.New("endpoint has no pod information")
	}

	ip := pod.GetIPAddress()
	if _, exists := emc.endpoints[ip]; exists {
		return fmt.Errorf("endpoint with IP %s already registered", ip)
	}

	// Create per-endpoint state
	epCtx, epCancel := context.WithCancel(emc.ctx)
	state := &endpointState{
		server:    ep,
		ctx:       epCtx,
		cancel:    epCancel,
		ticker:    time.NewTicker(emc.config.Interval),
		done:      make(chan struct{}),
		maxErrors: 5,
	}
	emc.endpoints[ip] = state

	if emc.running { // Start collection if collector is running
		emc.startEndpointCollectionLocked(state)
	}

	emc.logger.Info("registered endpoint for metrics collection", "endpoint", pod.String())
	return nil
}

// RemoveEndpoint removes an endpoint from collection.
func (emc *Collector) RemoveEndpoint(ep dltypes.Endpoint) error {
	emc.mu.Lock()
	defer emc.mu.Unlock()

	pod := ep.GetPod()
	ip := pod.GetIPAddress()
	state, exists := emc.endpoints[ip]
	if !exists {
		return fmt.Errorf("no server with IP %s registered", ip)
	}

	emc.stopEndpointCollectionLocked(state)

	delete(emc.endpoints, ip)
	emc.logger.Info("unregistered endpoint from metrics collection", "endpoint", pod.String())
	return nil
}

// startEndpointCollectionLocked starts metrics collection for an endpoint.
// As the name implies, this function must be called with collector lock held.
func (emc *Collector) startEndpointCollectionLocked(state *endpointState) {
	state.wg.Add(1)

	extractors := []dltypes.Extractor{}
	for _, extractor := range emc.extractors {
		extractors = append(extractors, extractor)
	}
	go emc.collectForEndpoint(state, extractors)

	emc.logger.Info("started collection goroutine for endpoint",
		"endpoint", state.server.GetPod().String())
}

// stopEndpointCollectionLocked stops collection for an endpoint.
// As the function name implies, it must be called with lock held.
func (emc *Collector) stopEndpointCollectionLocked(state *endpointState) {
	state.cancel()
	close(state.done)
	state.ticker.Stop()
	state.wg.Wait()
	emc.logger.Info("stopped collection goroutine for endpoint",
		"endpoint", state.server.GetPod().String())
}

// collectForEndpoint handles metrics collection for a single endpoint (runs in its own goroutine).
func (emc *Collector) collectForEndpoint(state *endpointState, extractors []dltypes.Extractor) {
	defer state.wg.Done()
	endpoint := state.server.GetPod().String()

	emc.logger.Info("starting metrics collection goroutine", "endpoint", endpoint)
	defer func() {
		emc.logger.Info("metrics collection goroutine finished", "endpoint", endpoint)
	}()

	for {
		select {
		case <-state.ctx.Done(): // endpoint-specific context cancelled
			return
		case <-emc.ctx.Done(): // Global collector context cancelled
			return
		case <-state.done: // Explicit stop signal
			return
		case <-state.ticker.C:
			if err := emc.collectMetrics(state, extractors); err != nil {
				state.errors++
				emc.logger.Error(err, "Failed to collect metrics", "endpoint", endpoint, "errors", state.errors)

				if state.errors >= state.maxErrors {
					emc.logger.Error(nil, "too many errors, stopping collection", "endpoint", endpoint)
					return
				}
			} else { // Reset error count on successful collection
				state.errors = 0
				state.lastUpdate = time.Now()
			}
		}
	}
}

// collectMetricsForServer performs the actual metrics collection for a server
func (smc *Collector) collectMetrics(state *endpointState, _ []dltypes.Extractor) error {
	pod := state.server.GetPod()
	if pod == nil {
		return errors.New("no pod information available for server")
	}

	// get the pool for the port
	// create a request via the (TBD) http client (request with cancel, timeout, ...)
	// send to the extractors on success

	return nil
}
