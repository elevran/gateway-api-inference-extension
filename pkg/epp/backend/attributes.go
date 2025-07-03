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
	"sync"
)

// Cloneable types support cloning of the value.
type Cloneable interface {
	Clone() Cloneable
}

// attributes is used to store flexible, cloneable metadata or traits
// across different aspects of a ModelServer.
type attributes struct {
	mu   sync.RWMutex
	data map[string]Cloneable
}

// newAttributes() return a new attribute store
func newAttributes() *attributes {
	return &attributes{
		data: make(map[string]Cloneable),
	}
}

// put a new attribute into the store
func (a *attributes) put(key string, value Cloneable) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.data[key] = value // TODO: Clone into map?
}

// get an attribute from the store
func (a *attributes) get(key string) (Cloneable, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	val, ok := a.data[key]
	if !ok {
		return nil, false
	}
	return val.Clone(), true
}

// keys returns an array of all the names of attributes stored
func (a *attributes) keys() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	keys := make([]string, 0, len(a.data))
	for k := range a.data {
		keys = append(keys, k)
	}
	return keys
}

// Clone the attributes struct itself
func (a *attributes) Clone() *attributes {
	a.mu.RLock()
	defer a.mu.RUnlock()

	m := make(map[string]Cloneable, len(a.data))
	for k, v := range a.data {
		m[k] = v.Clone()
	}
	return &attributes{
		data: m,
	}
}
