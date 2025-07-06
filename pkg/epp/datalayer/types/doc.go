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

// TODO
// To avoid circular dependency store basic types in this directory
// and complex implementations in the root directory (including those that
// depend on backend/metrics package during the refactor; (e.g., Endpoint
// implemented via backend/metrics.PodMetrics)).
// An alternative would have been to create an explicit `impl`` package under
// datalayer and move the implementations there while keeping interfaces at
// the root. The root would then.
// At the end of the day, selecting types or impl is a matter of preference and
// consistency - either works...
