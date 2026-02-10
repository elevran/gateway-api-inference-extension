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
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// EventType identifies the kind of k8s object mutation that triggered a notification.
type EventType int

const (
	// EventAddOrUpdate is fired when a k8s object is created or updated.
	EventAddOrUpdate EventType = iota
	// EventDelete is fired when a k8s object is deleted.
	EventDelete
)

// NotificationEvent carries the event type and the affected object.
// Object is deep-copied by the framework core before delivery.
type NotificationEvent struct {
	Type EventType
	// Object is the current state of the object (for add/update) or the
	// last known state (for delete).
	Object *unstructured.Unstructured
}

// NotificationSource is a DataSource that is event-driven rather than poll-based.
// It watches a single k8s GVK. The framework core owns the cache informer and
// calls Notify on events; the source dispatches to its extractors.
//
// One source per GVK; registration is exclusive via the DataSourceRegistry.
// Collect is a no-op — all data flows through Notify.
type NotificationSource interface {
	DataSource
	// GVK returns the GroupVersionKind this source watches.
	GVK() schema.GroupVersionKind
	// Notify is called by the framework core when an informer event fires.
	// The event object is already deep-copied.
	Notify(ctx context.Context, event NotificationEvent)
}

// NotificationExtractor processes k8s object events pushed from a
// NotificationSource. It embeds Extractor so it can be stored via
// DataSource.AddExtractor. The Extractor.Extract method is never called
// on the notification path — ExtractNotification is used instead.
type NotificationExtractor interface {
	Extractor
	// ExtractNotification processes a notification event. Called synchronously
	// by the source in event order.
	ExtractNotification(ctx context.Context, event NotificationEvent) error
}

// UnstructuredType is the reflect.Type for unstructured.Unstructured,
// used by notification extractors to declare their expected input type.
var UnstructuredType = reflect.TypeOf(unstructured.Unstructured{})
