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

package logging

import (
	"os"

	"github.com/go-logr/logr"
)

// Fatal calls logger.Error followed by os.Exit(1).
//
// This is a utility function and should not be used in production code!
func Fatal(logger logr.Logger, err error, msg string, keysAndValues ...any) {
	logger.Error(err, msg, keysAndValues...)
	os.Exit(1)
}
