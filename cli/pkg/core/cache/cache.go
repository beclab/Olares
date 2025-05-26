/*
 Copyright 2021 The KubeSphere Authors.

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

package cache

import "sync"

type Cache struct {
	store sync.Map
}

func NewCache() *Cache {
	var m Cache
	return &m
}

func (c *Cache) Set(k string, v interface{}) {
	c.store.Store(k, v)
}

// GetOrSet returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (c *Cache) GetOrSet(k string, v interface{}) (interface{}, bool) {
	return c.store.LoadOrStore(k, v)
}

func (c *Cache) Get(k string) (interface{}, bool) {
	return c.store.Load(k)
}

func (c *Cache) Range(f func(key, value interface{}) bool) {
	c.store.Range(f)
}

func (c *Cache) Delete(k string) {
	c.store.Delete(k)
}

func (c *Cache) Clean() {
	c.store.Range(func(key, value interface{}) bool {
		c.store.Delete(key)
		return true
	})
}

func (c *Cache) GetMustInt(k string) (int, bool) {
	v, ok := c.Get(k)
	res, assert := v.(int)
	if !assert {
		return res, false
	}
	return res, ok
}

func (c *Cache) GetMustString(k string) (string, bool) {
	v, ok := c.Get(k)
	res, assert := v.(string)
	if !assert {
		return res, false
	}
	return res, ok
}

func (c *Cache) GetMustBool(k string) (bool, bool) {
	v, ok := c.Get(k)
	res, assert := v.(bool)
	if !assert {
		return res, false
	}
	return res, ok
}
