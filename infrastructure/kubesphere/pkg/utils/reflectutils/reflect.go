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

package reflectutils

import (
	"reflect"
)

func Override(left interface{}, right interface{}) {
	if reflect.ValueOf(left).IsNil() || reflect.ValueOf(right).IsNil() {
		return
	}

	if reflect.ValueOf(left).Type().Kind() != reflect.Ptr ||
		reflect.ValueOf(right).Type().Kind() != reflect.Ptr ||
		reflect.ValueOf(left).Kind() != reflect.ValueOf(right).Kind() {
		return
	}

	oldVal := reflect.ValueOf(left).Elem()
	newVal := reflect.ValueOf(right).Elem()

	for i := 0; i < oldVal.NumField(); i++ {
		val := newVal.Field(i).Interface()
		if !reflect.DeepEqual(val, reflect.Zero(reflect.TypeOf(val)).Interface()) {
			oldVal.Field(i).Set(reflect.ValueOf(val))
		}
	}
}
