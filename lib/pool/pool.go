// Copyright 2023 Huawei Cloud Computing Technologies Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pool

import (
	"sync"

	"github.com/openGemini/openGemini/lib/bufferpool"
	"github.com/openGemini/openGemini/lib/cpu"
)

type HitRatioHook struct {
	IncrTotal func(int64)
	IncrHit   func(int64)
}

func NewHitRatioHook(total, hit func(int64)) *HitRatioHook {
	return &HitRatioHook{
		IncrTotal: total,
		IncrHit:   hit,
	}
}

type FixedPool struct {
	pool chan interface{}
	new  func() interface{}
	hook *HitRatioHook
}

func (p *FixedPool) Reset(size int, new func() interface{}, hook *HitRatioHook) {
	p.pool = make(chan interface{}, size)
	p.new = new
	p.hook = hook
}

func (p *FixedPool) newObject() interface{} {
	if p.new == nil {
		return nil
	}
	return p.new()
}

func (p *FixedPool) Get() interface{} {
	if p.hook != nil {
		p.hook.IncrTotal(1)
	}
	select {
	case iw := <-p.pool:
		if p.hook != nil {
			p.hook.IncrHit(1)
		}
		return iw
	default:
		break
	}
	return p.newObject()
}

func (p *FixedPool) Put(v interface{}) {
	select {
	case p.pool <- v:
	default:
		break
	}
}

type Object interface {
	Instance() Object
	MemSize() int
}

type ObjectPool struct {
	pool  sync.Pool
	local chan Object
	obj   Object
	hook  *HitRatioHook

	maxLocalCacheSize int
}

func NewObjectPool(size int, obj Object, maxLocalCacheSize int) *ObjectPool {
	if size <= 0 || size > bufferpool.MaxLocalCacheLen {
		size = bufferpool.MaxLocalCacheLen
	}

	return &ObjectPool{
		pool:              sync.Pool{},
		local:             make(chan Object, size),
		obj:               obj,
		maxLocalCacheSize: maxLocalCacheSize,
	}
}

func (p *ObjectPool) SetHitRatioHook(hook *HitRatioHook) {
	p.hook = hook
}

func (p *ObjectPool) stat(hit bool) {
	if p.hook == nil {
		return
	}

	p.hook.IncrTotal(1)
	if hit {
		p.hook.IncrHit(1)
	}
}

func (p *ObjectPool) Get() Object {
	var obj Object = nil
	select {
	case obj = <-p.local:
		break
	default:
		o, ok := p.pool.Get().(Object)
		if ok {
			obj = o
		}
	}

	if obj == nil {
		p.stat(false)
		return p.obj.Instance()
	}

	p.stat(true)
	return obj
}

func (p *ObjectPool) Put(v Object) {
	if v.MemSize() > p.maxLocalCacheSize {
		p.pool.Put(v)
		return
	}

	select {
	case p.local <- v:
	default:
		p.pool.Put(v)
	}
}

type FixedCachePool struct {
	cache chan interface{}
	pool  sync.Pool
}

func NewFixedCachePool() *FixedCachePool {
	n := cpu.GetCpuNum() * 2
	if n < 4 {
		n = 4
	}
	if n > 256 {
		n = 256
	}
	return &FixedCachePool{
		cache: make(chan interface{}, n),
	}
}

func (p *FixedCachePool) Reset(size int) {
	p.cache = make(chan interface{}, size)
}

func (p *FixedCachePool) Get() interface{} {
	select {
	case iw := <-p.cache:
		return iw
	default:
		return p.pool.Get()
	}
}

func (p *FixedCachePool) Put(v interface{}) {
	select {
	case p.cache <- v:
	default:
		p.pool.Put(v)
	}
}

// FixedPoolV2 is a fixed size pool with object generic type information
type FixedPoolV2[T any] struct {
	pool chan T
	new  func() T
}

func NewFixedPoolV2[T any](new func() T, size int) *FixedPoolV2[T] {
	return &FixedPoolV2[T]{
		pool: make(chan T, size),
		new:  new,
	}
}

func (p *FixedPoolV2[T]) Get() T {
	select {
	case item := <-p.pool:
		return item
	default:
		return p.new()
	}
}

func (p *FixedPoolV2[T]) Put(item T) {
	select {
	case p.pool <- item:
	default:
	}
}

func (p *FixedPoolV2[T]) Reset(size int, new func() T) {
	p.pool = make(chan T, size)
	p.new = new
}

func (p *FixedPoolV2[T]) Len() int {
	return len(p.pool)
}

var intSlicePool sync.Pool

func GetIntSlice(size int) []int {
	v, ok := intSlicePool.Get().(*[]int)
	if !ok || v == nil || cap(*v) < size {
		return make([]int, size)
	}
	return (*v)[:size]
}

func PutIntSlice(v []int) {
	intSlicePool.Put(&v)
}
