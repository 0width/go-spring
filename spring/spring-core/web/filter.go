/*
 * Copyright 2012-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package web

import (
	"regexp"
)

// Filter 过滤器接口，Invoke 通过 chain.Next() 驱动链条向后执行。
type Filter interface {
	Invoke(ctx Context, chain FilterChain)
}

// funcFilter 封装 func 形式的过滤器。
type funcFilter struct {
	f func(ctx Context, chain FilterChain)
	p []string
}

// FuncFilter 封装 func 形式的过滤器。
func FuncFilter(f func(ctx Context, chain FilterChain)) *funcFilter {
	return &funcFilter{f: f}
}

// URLPatterns 为过滤器设置 url 匹配模式。
func (f *funcFilter) URLPatterns(p ...string) *funcFilter {
	f.p = p
	return f
}

func (f *funcFilter) Invoke(ctx Context, chain FilterChain) {
	f.f(ctx, chain)
}

// handlerFilter 包装 Web 处理接口的过滤器
type handlerFilter struct {
	fn Handler
}

// HandlerFilter 把 Web 处理接口转换成过滤器。
func HandlerFilter(fn Handler) Filter {
	return &handlerFilter{fn: fn}
}

func (h *handlerFilter) Invoke(ctx Context, _ FilterChain) {
	h.fn.Invoke(ctx)
}

// FilterChain 过滤器链条接口
type FilterChain interface {

	// Next 立即执行下一个节点。
	Next(ctx Context)

	// Continue 结束当前节点后执行下一个节点，用于减少调用栈深度。
	Continue(ctx Context)
}

// DefaultFilterChain 默认的过滤器链条
type DefaultFilterChain struct {
	filters  []Filter
	index    int
	lazyNext bool
}

// NewDefaultFilterChain DefaultFilterChain 的构造函数
func NewDefaultFilterChain(filters []Filter) *DefaultFilterChain {
	return &DefaultFilterChain{filters: filters}
}

func (chain *DefaultFilterChain) next(ctx Context) {
	f := chain.filters[chain.index]
	chain.index++
	f.Invoke(ctx, chain)
}

func (chain *DefaultFilterChain) Next(ctx Context) {
	if chain.index < len(chain.filters) {
		chain.next(ctx)
		for chain.lazyNext {
			chain.lazyNext = false
			chain.next(ctx)
		}
	}
}

func (chain *DefaultFilterChain) Continue(ctx Context) {
	chain.lazyNext = true
}

type urlPatterns struct {
	m map[*regexp.Regexp][]Filter
}

func (p *urlPatterns) Get(path string) []Filter {
	for pattern, filters := range p.m {
		if pattern.MatchString(path) {
			return filters
		}
	}
	return nil
}

// URLPatterns 根据 Filter 的 URL 匹配表达式进行分组。
func URLPatterns(filters []Filter) (*urlPatterns, error) {

	filterMap := make(map[string][]Filter)
	for _, filter := range filters {
		var patterns []string
		if p, ok := filter.(interface{ URLPatterns() []string }); ok {
			patterns = p.URLPatterns()
		} else {
			patterns = []string{"/*"}
		}
		for _, pattern := range patterns {
			filterMap[pattern] = append(filterMap[pattern], filter)
		}
	}

	filterPatterns := make(map[*regexp.Regexp][]Filter)
	for pattern, filter := range filterMap {
		exp, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		filterPatterns[exp] = filter
	}
	return &urlPatterns{m: filterPatterns}, nil
}
