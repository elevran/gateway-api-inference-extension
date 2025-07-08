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
	"context"
	"reflect"
)

// DataSource is an interface required from all datalayer data collection
// sources.
type DataSource interface {
	// Name returns the name of this datasource
	Name() string

	// Start begins the collection process.
	Start(ctx context.Context) error

	// Stop stops the collection process.
	Stop() error

	// OutputType returns the type of information collected.
	// TODO: this could be made private and excluded from the interface.
	//       Use will likely be in the context of debugging/logging.
	OutputType() reflect.Type

	// AddExtractor adds an extractor to the data source.
	// The extractor will be called whenever the data source might
	// have some new raw information regarding an endpoint.
	// The Extractor's expected input type is validated when it is
	// registered
	AddExtractor(extractor Extractor) error

	// TODO: the following is useful for a data source that operates on
	// endpoints and might not be relevant for "global/system" collectors which
	// might not need the concept of an endpoint. This can be split, if needed,
	// to a separate interface in the future.

	// AddEndpoint adds an endpoint to collect from.
	AddEndpoint(ep Endpoint) error

	// RemoveEndpoint removes an endpoint from collection.
	RemoveEndpoint(ep Endpoint) error
}

type Extractor interface {
	// Name returns the name of the extractor
	Name() string

	// ExpectedType defines the type expected by the extractor. It must match
	// the DataSource.OutputType() the extractor registers for.
	ExpectedType() reflect.Type

	// Extract transforms the data source output into a concrete attribute that
	// is stored on the given endpoint.
	Extract(data any, ep Endpoint)
}
