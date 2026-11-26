/*
Copyright 2018 The KubeSphere Authors.

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

// Package deep provides function deep.Equal which is like reflect.DeepEqual but
// returns a list of differences. This is helpful when comparing complex types
// like structures and maps.
package reflectutils

import (
	"errors"
	"reflect"
)

var (
	// FloatPrecision is the number of decimal places to round float values
	// to when comparing.
	FloatPrecision = 10

	// MaxDiff specifies the maximum number of differences to return.
	MaxDiff = 10

	// MaxDepth specifies the maximum levels of a struct to recurse into,
	// if greater than zero. If zero, there is no limit.
	MaxDepth = 0

	// LogErrors causes errors to be logged to STDERR when true.
	LogErrors = false

	// CompareUnexportedFields causes unexported struct fields, like s in
	// T{s int}, to be compared when true.
	CompareUnexportedFields = false
)

var (
	// ErrMaxRecursion is logged when MaxDepth is reached.
	ErrMaxRecursion = errors.New("recursed to MaxDepth")

	// ErrTypeMismatch is logged when Equal passed two different types of values.
	ErrTypeMismatch = errors.New("variables are different reflect.Type")

	// ErrNotHandled is logged when a primitive Go kind is not handled.
	ErrNotHandled = errors.New("cannot compare the reflect.Kind")
)

type cmp struct {
	diff        []string
	buff        []string
	floatFormat string
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()
