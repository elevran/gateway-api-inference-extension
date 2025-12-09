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

package options

import (
	"flag"
	"fmt"
	"io"
	"time"
)

// Flag defines parameters needed to manage command line flags.
type Flag struct {
	Name       string // CLI flag name.
	DefValue   any    // default value, required (to ensure Flag value type is defined).
	Usage      string // help text.
	Deprecated bool   // optional mark as deprecated.
	ReplacedBy string // optional replacement message.
}

// AddFlags registers a list of Flag definitions with a FlagSet (defaulting to
// flag.CommandLine if unspecified)), binding them to the pointer variables provided
// in the vars map.
// TODO: generics
func AddFlags(fs *flag.FlagSet, flags []Flag, vars map[string]any) error {
	if len(flags) != len(vars) {
		return fmt.Errorf("mismatch flags (%d) and vars (%d) count", len(flags), len(vars))
	}

	if fs == nil {
		fs = flag.CommandLine
	}

	for _, f := range flags {
		if f.DefValue == nil { // a default value is required to determine types
			return fmt.Errorf("flag %q must have a non-nil default value", f.Name)
		}

		ptr, ok := vars[f.Name]
		// TODO: validate ptr is indeed a pointer type?
		if !ok { // no destination variable
			return fmt.Errorf("variable pointer for flag %q not provided", f.Name)
		}

		switch def := f.DefValue.(type) {
		case string:
			p, ok := ptr.(*string)
			if !ok {
				return typeError(f.Name, ptr, "string")
			}
			*p = def
			fs.StringVar(p, f.Name, def, f.Usage)
		case int:
			p, ok := ptr.(*int)
			if !ok {
				return typeError(f.Name, ptr, "int")
			}
			*p = def
			fs.IntVar(p, f.Name, def, f.Usage)
		case bool:
			p, ok := ptr.(*bool)
			if !ok {
				return typeError(f.Name, ptr, "bool")
			}
			*p = def
			fs.BoolVar(p, f.Name, def, f.Usage)
		case time.Duration:
			p, ok := ptr.(*time.Duration)
			if !ok {
				return typeError(f.Name, ptr, "time.Duration")
			}
			*p = def
			fs.DurationVar(p, f.Name, def, f.Usage)
		default:
			return fmt.Errorf("unsupported flag type for %q: %T", f.Name, def)
		}

		if f.Deprecated { // wrap the value with deprecation warning
			fl := fs.Lookup(f.Name)
			if fl == nil {
				return fmt.Errorf("failed to lookup deprecated flag %q in set", f.Name)
			}
			fl.Value = &deprecatedValue{
				Value:      fl.Value,
				name:       f.Name,
				writer:     fs.Output(),
				replacedBy: f.ReplacedBy,
			}
		}
	}

	return nil
}

// deprecatedValue wraps a standard flag.Value to inject a warning message
// when the deprecated flag is used.
type deprecatedValue struct {
	flag.Value
	warned     bool
	name       string
	replacedBy string
	writer     io.Writer
}

// Set is called when the flag is parsed from the command line.
func (d *deprecatedValue) Set(s string) error {
	err := d.Value.Set(s) // delegate to the flag.Value

	if err == nil && !d.warned {
		d.warned = true
		if d.replacedBy != "" {
			fmt.Fprintf(d.writer, "Warning: --%s is deprecated; use %s instead.\n", d.name, d.replacedBy)
		} else {
			fmt.Fprintf(d.writer, "Warning: --%s is deprecated and will be removed in an upcoming release.\n", d.name)
		}
	}
	return err
}

// typeError creates a clear error message for flag type mismatches.
func typeError(name string, got any, expected string) error {
	return fmt.Errorf("flag %q: variable must be *%s, got %T", name, expected, got)
}

// TODO: generics GetFlagValue[T]
// GetStringFlagValue retrieves the current value (default or set) of a string flag
// by name from the specified FlagSet (or flag.CommandLine if nil).
func GetStringFlagValue(fs *flag.FlagSet, name string) (string, error) {
	if fs == nil {
		fs = flag.CommandLine
	}

	f := fs.Lookup(name)
	if f == nil {
		return "", fmt.Errorf("flag not found: %s", name)
	}
	val, ok := f.Value.(flag.Getter)
	if !ok {
		return "", fmt.Errorf("flag %s value does not support flag.Getter interface", name)
	}
	underlying := val.Get()
	strptr, ok := underlying.(*string)
	if !ok {
		return "", fmt.Errorf("flag %s is not a string type, got %T", name, underlying)
	}
	return *strptr, nil
}

// GetBoolFlagValue retrieves the current value (default or set) of a boolean flag
// by name from the specified FlagSet (or flag.CommandLine if nil).
func GetBoolFlagValue(fs *flag.FlagSet, name string) (bool, error) {
	if fs == nil {
		fs = flag.CommandLine
	}

	f := fs.Lookup(name)
	if f == nil {
		return false, fmt.Errorf("flag not found: %s", name)
	}
	val, ok := f.Value.(flag.Getter)
	if !ok {
		return false, fmt.Errorf("flag %s value does not support flag.Getter interface", name)
	}
	underlying := val.Get()
	boolptr, ok := underlying.(*bool)
	if !ok {
		return false, fmt.Errorf("flag %s is not a bool type, got %T", name, underlying)
	}
	return *boolptr, nil
}

/*
package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"time"
	"io"
)

// User-Facing Flag Definition

type Flag struct {
	Name       string
	Default    any
	Usage      string
	Deprecated bool
	ReplacedBy string
}

// Deprecation Wrapper (Unchanged)

type deprecatedValue struct {
	flag.Value
	writer     io.Writer
	name       string
	replacedBy string
	warned     bool
}

func (d *deprecatedValue) Set(s string) error {
	err := d.Value.Set(s)
	if err == nil && !d.warned {
		d.warned = true
		if d.replacedBy != "" {
			fmt.Fprintf(d.writer,
				"Warning: --%s is deprecated; use --%s instead\n",
				d.name, d.replacedBy)
		} else {
			fmt.Fprintf(d.writer,
				"Warning: --%s is deprecated\n", d.name)
		}
	}
	return err
}

// AddFlags Implementation (Improved with Generics)
// setGenericFlag sets the default value using reflection and registers the flag
// using the appropriate type-specific flag function.
// This function replaces the large switch statement in the original AddFlags.
func setGenericFlag[T any](fs *flag.FlagSet, name string, ptr any, def T, usage string) error {
	// 1. Assert the pointer type
	p, ok := ptr.(*T)
	if !ok {
		// Use reflect.TypeOf to get the expected type string for the error message
		expectedType := reflect.TypeOf((*T)(nil)).Elem().String()
		return fmt.Errorf("flag %q: variable must be *%s, got %T", name, expectedType, ptr)
	}

	// 2. Set the default value
	*p = def

	// 3. Register the flag using the appropriate function (still needs a switch
	//    because flag package functions are not generic, but this switch is simpler)
	switch any(p).(type) {
	case *string:
		fs.StringVar(p.(*string), name, def.(string), usage)
	case *int:
		fs.IntVar(p.(*int), name, def.(int), usage)
	case *bool:
		fs.BoolVar(p.(*bool), name, def.(bool), usage)
	case *float64:
		fs.Float64Var(p.(*float64), name, def.(float64), usage)
	case *time.Duration:
		fs.DurationVar(p.(*time.Duration), name, def.(time.Duration), usage)
	default:
		// This should theoretically not be hit if AddFlags calls this correctly
		return fmt.Errorf("unsupported generic type for registration: %T", def)
	}

	return nil
}

// AddFlags registers a list of Flag definitions with a FlagSet using the generic helper.
func AddFlags(fs *flag.FlagSet, flags []Flag, vars map[string]any) error {
	for _, f := range flags {
		ptr, ok := vars[f.Name]
		if !ok {
			return fmt.Errorf("AddFlags: no variable provided for flag %q", f.Name)
		}

		// Use a switch on the concrete type of the default value
		// and delegate the binding and registration to the generic helper.
		var err error
		switch def := f.Default.(type) {
		case string:
			err = setGenericFlag(fs, f.Name, ptr, def, f.Usage)
		case int:
			err = setGenericFlag(fs, f.Name, ptr, def, f.Usage)
		case bool:
			err = setGenericFlag(fs, f.Name, ptr, def, f.Usage)
		case float64:
			err = setGenericFlag(fs, f.Name, ptr, def, f.Usage)
		case time.Duration:
			err = setGenericFlag(fs, f.Name, ptr, def, f.Usage)
		default:
			return fmt.Errorf("unsupported flag type for %q: %T", f.Name, def)
		}

        if err != nil {
            return err
        }

		// Insert deprecation wrapper *after* registration (Unchanged)
		if f.Deprecated {
			fl := fs.Lookup(f.Name)
			if fl != nil {
				fl.Value = &deprecatedValue{
					Value:      fl.Value,
					writer:     fs.Output(),
					name:       f.Name,
					replacedBy: f.ReplacedBy,
				}
			}
		}
	}

	return nil
}

// Flag Retrieval (Improved with Generics)
// GetFlagValue retrieves the current value (default or set) of any flag by name.
// It uses a type parameter T to ensure the value is returned in the expected type
// and enforces type safety at compile time.
func GetFlagValue[T any](fs *flag.FlagSet, name string) (T, error) {
	// Idiomatic normalization: If fs is nil, use the default global FlagSet.
	if fs == nil {
		fs = flag.CommandLine
	}

	// 1. Look up the flag definition
	f := fs.Lookup(name)
	if f == nil {
		var zero T
		return zero, fmt.Errorf("flag not found: %s", name)
	}

	// 2. Use flag.Getter to safely extract the underlying value
	val, ok := f.Value.(flag.Getter)
	if !ok {
		var zero T
		return zero, fmt.Errorf("flag %s value does not support flag.Getter interface", name)
	}

	underlyingValue := val.Get()

	// 3. Assert that the underlying value is a pointer to the generic type T (*T)
	// We use reflection here because we need to dynamically check the type of the pointer.
	// We are looking for a *T, not a T.

	// Check if the underlying value is a pointer
	if reflect.TypeOf(underlyingValue).Kind() != reflect.Ptr {
		var zero T
		return zero, fmt.Errorf("internal flag error: storage for flag %s is not a pointer, got %T", name, underlyingValue)
	}

	// Check if the pointer's element type matches the requested type T
	expectedType := reflect.TypeOf((*T)(nil)).Elem()
	actualElementType := reflect.TypeOf(underlyingValue).Elem()

	if actualElementType != expectedType {
		var zero T
		return zero, fmt.Errorf("flag %s type mismatch: requested %s, but flag is stored as %s",
			name, expectedType, actualElementType)
	}

	// 4. Safely assert the value to the correct pointer type (*T) and dereference it
	// We must first assert it back to the generic pointer type, then dereference.
	// This uses a type assertion on the underlying interface value.
	ptrToT, ok := underlyingValue.(*T)
	if !ok {
		// This should not happen if the previous type checks passed, but is a safety net.
		var zero T
		return zero, fmt.Errorf("internal type error for flag %s: cannot assert to *%s", name, expectedType)
	}

	return *ptrToT, nil
}


// Example main
func main() {
	// 1. Define Flags using the new structure
	flags := []Flag{
		{Name: "server-port", Default: 8080, Usage: "server port"},
		{Name: "debug-mode", Default: false, Usage: "enable debug"},
		{Name: "request-timeout", Default: time.Second * 3, Usage: "timeout"},
		{
			Name:       "old-timeout",
			Default:    time.Second * 30,
			Usage:      "DEPRECATED",
			Deprecated: true,
			ReplacedBy: "request-timeout",
		},
		{Name: "host-name", Default: "localhost", Usage: "host name"},
	}

	// 2. Define user variables (pointers) to bind flags into
	var portVal int
	var debugVal bool
	var timeoutVal time.Duration
	var oldTimeoutVal time.Duration
	var hostVal string

	vars := map[string]any{
		"server-port":     &portVal,
		"debug-mode":      &debugVal,
		"request-timeout": &timeoutVal,
		"old-timeout":     &oldTimeoutVal,
		"host-name":       &hostVal,
	}

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	if err := AddFlags(fs, flags, vars); err != nil {
		panic(err)
	}

    // Set some runtime values before parsing (simulating command line)
    fs.Set("server-port", "9000")
    fs.Set("debug-mode", "true")

	// fs.Parse(os.Args[1:]) // Typically parse here, but we set values manually for this example

	fmt.Println("--- 1. Verification of Bound Values (via Pointers) ---")
	fmt.Printf("server-port (int bound): %d\n", portVal) // Should be 9000
	fmt.Printf("debug-mode (bool bound): %t\n", debugVal) // Should be true

	// 3. Retrieve values using the new generic GetFlagValue function
	fmt.Println("\n--- 2. Retrieving Values via Generic Function ---")

	// Retrieve int
	port, err := GetFlagValue[int](fs, "server-port")
	if err == nil {
		fmt.Printf("Retrieved int: 'server-port' = %d\n", port)
	}

	// Retrieve bool
	debug, err := GetFlagValue[bool](fs, "debug-mode")
	if err == nil {
		fmt.Printf("Retrieved bool: 'debug-mode' = %t\n", debug)
	}

	// Retrieve time.Duration (default value)
	timeout, err := GetFlagValue[time.Duration](fs, "request-timeout")
	if err == nil {
		fmt.Printf("Retrieved time.Duration: 'request-timeout' = %s\n", timeout)
	}

	// --- Error Testing ---

	// Test type mismatch (Requesting string for an int flag)
	_, err = GetFlagValue[string](fs, "server-port")
	if err != nil {
		fmt.Println("\n--- 3. Error Testing ---")
		fmt.Printf("Type Mismatch Error (Correctly caught): %v\n", err)
	}
}
*/
