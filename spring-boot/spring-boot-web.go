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

package SpringBoot

import (
	"errors"
	"reflect"
	"runtime"
	"strings"

	"github.com/go-spring/go-spring-web/spring-web"
	"github.com/go-spring/go-spring/spring-core"
)

type MethodHandler struct {
	Receiver   reflect.Type
	MethodName string
}

// WebMapping Web 路由映射表
type WebMapping struct {
	Mappings map[string]*Mapping
}

// NewWebMapping WebMapping 的构造函数
func NewWebMapping() *WebMapping {
	return &WebMapping{
		Mappings: make(map[string]*Mapping),
	}
}

// Request 路由注册
func (m *WebMapping) Request(method uint32, path string, fn interface{}, filters []SpringWeb.Filter) *Mapping {
	var handler interface{}

	fnType := reflect.TypeOf(fn)
	fnValue := reflect.ValueOf(fn)

	if fnType.AssignableTo(SpringWeb.FnHandlerType) { // 可直接赋值
		v := fnValue.Convert(SpringWeb.FnHandlerType)
		v = v.Convert(SpringWeb.HandlerType)
		handler = v.Interface().(SpringWeb.Handler)

	} else if fnType.AssignableTo(SpringWeb.HandlerType) { // 可直接赋值
		v := fnValue.Convert(SpringWeb.HandlerType)
		handler = v.Interface().(SpringWeb.Handler)

	} else {
		fnPtr := fnValue.Pointer()
		fnInfo := runtime.FuncForPC(fnPtr)
		s := strings.Split(fnInfo.Name(), "/")
		ss := strings.Split(s[len(s)-1], ".")

		// 获取方法名称
		var methodName string
		if len(ss) == 3 {
			methodName = ss[2]
		} else {
			panic(errors.New("error method func"))
		}

		handler = &MethodHandler{
			Receiver:   fnType.In(0),
			MethodName: methodName,
		}
	}

	mapping := newMapping(method, path, handler, filters)
	m.Mappings[mapping.Key()] = mapping
	return mapping
}

// Mapping 封装 Web 路由映射
type Mapping struct {
	handler interface{}
	mapper  *SpringWeb.Mapper       // 路由映射器
	ports   []int                   // 路由端口
	cond    *SpringCore.Conditional // 判断条件
}

// newMapping Mapping 的构造函数
func newMapping(method uint32, path string, handler interface{}, filters []SpringWeb.Filter) *Mapping {
	return &Mapping{
		handler: handler,
		mapper:  SpringWeb.NewMapper(method, path, nil, filters),
		cond:    SpringCore.NewConditional(),
	}
}

// Mapper 返回封装的 Mapper 对象
func (m *Mapping) Mapper() *SpringWeb.Mapper {
	return m.mapper
}

// Key 返回 Mapper 的标识符
func (m *Mapping) Key() string {
	return m.mapper.Key()
}

// Method 返回 Mapper 的方法
func (m *Mapping) Method() uint32 {
	return m.mapper.Method()
}

// Path 返回 Mapper 的路径
func (m *Mapping) Path() string {
	return m.mapper.Path()
}

// HandlerSelector 返回处理函数选择器
func (m *Mapping) Handler() interface{} {
	return m.handler
}

// Ports 返回路由期望的端口
func (m *Mapping) Ports() []int {
	return m.ports
}

// OnPorts 设置路由期望的端口
func (m *Mapping) OnPorts(ports ...int) *Mapping {
	m.ports = ports
	return m
}

// Filters 返回 Mapper 的过滤器列表
func (m *Mapping) Filters() []SpringWeb.Filter {
	return m.mapper.Filters()
}

// Or c=a||b
func (m *Mapping) Or() *Mapping {
	m.cond.Or()
	return m
}

// And c=a&&b
func (m *Mapping) And() *Mapping {
	m.cond.And()
	return m
}

// ConditionOn 设置一个 Condition
func (m *Mapping) ConditionOn(cond SpringCore.Condition) *Mapping {
	m.cond.OnCondition(cond)
	return m
}

// ConditionNot 设置一个取反的 Condition
func (m *Mapping) ConditionNot(cond SpringCore.Condition) *Mapping {
	m.cond.OnConditionNot(cond)
	return m
}

// ConditionOnProperty 设置一个 PropertyCondition
func (m *Mapping) ConditionOnProperty(name string) *Mapping {
	m.cond.OnProperty(name)
	return m
}

// ConditionOnMissingProperty 设置一个 MissingPropertyCondition
func (m *Mapping) ConditionOnMissingProperty(name string) *Mapping {
	m.cond.OnMissingProperty(name)
	return m
}

// ConditionOnPropertyValue 设置一个 PropertyValueCondition
func (m *Mapping) ConditionOnPropertyValue(name string, havingValue interface{}) *Mapping {
	m.cond.OnPropertyValue(name, havingValue)
	return m
}

// ConditionOnBean 设置一个 BeanCondition
func (m *Mapping) ConditionOnBean(selector interface{}) *Mapping {
	m.cond.OnBean(selector)
	return m
}

// ConditionOnMissingBean 设置一个 MissingBeanCondition
func (m *Mapping) ConditionOnMissingBean(selector interface{}) *Mapping {
	m.cond.OnMissingBean(selector)
	return m
}

// ConditionOnExpression 设置一个 ExpressionCondition
func (m *Mapping) ConditionOnExpression(expression string) *Mapping {
	m.cond.OnExpression(expression)
	return m
}

// ConditionOnMatches 设置一个 FunctionCondition
func (m *Mapping) ConditionOnMatches(fn SpringCore.ConditionFunc) *Mapping {
	m.cond.OnMatches(fn)
	return m
}

// ConditionOnProfile 设置一个 ProfileCondition
func (m *Mapping) ConditionOnProfile(profile string) *Mapping {
	m.cond.OnProfile(profile)
	return m
}

// Matches 成功返回 true，失败返回 false
func (m *Mapping) Matches(ctx SpringCore.SpringContext) bool {
	return m.cond.Matches(ctx)
}

// Swagger 生成并返回 Swagger 操作节点
func (m *Mapping) Swagger() *SpringWeb.Operation {
	return m.mapper.Swagger("")
}

// Router 路由分组
type Router struct {
	mapping  *WebMapping
	basePath string
	filters  []SpringWeb.Filter
	ports    []int                   // 路由端口
	cond     *SpringCore.Conditional // 判断条件
}

// newRouter Router 的构造函数
func newRouter(mapping *WebMapping, basePath string, filters []SpringWeb.Filter) *Router {
	return &Router{
		mapping:  mapping,
		basePath: basePath,
		filters:  filters,
		cond:     SpringCore.NewConditional(),
	}
}

// OnPorts 设置路由期望的端口
func (r *Router) OnPorts(ports ...int) *Router {
	r.ports = ports
	return r
}

// Or c=a||b
func (r *Router) Or() *Router {
	r.cond.Or()
	return r
}

// And c=a&&b
func (r *Router) And() *Router {
	r.cond.And()
	return r
}

// ConditionOn 设置一个 Condition
func (r *Router) ConditionOn(cond SpringCore.Condition) *Router {
	r.cond.OnCondition(cond)
	return r
}

// ConditionNot 设置一个取反的 Condition
func (r *Router) ConditionNot(cond SpringCore.Condition) *Router {
	r.cond.OnConditionNot(cond)
	return r
}

// ConditionOnProperty 设置一个 PropertyCondition
func (r *Router) ConditionOnProperty(name string) *Router {
	r.cond.OnProperty(name)
	return r
}

// ConditionOnMissingProperty 设置一个 MissingPropertyCondition
func (r *Router) ConditionOnMissingProperty(name string) *Router {
	r.cond.OnMissingProperty(name)
	return r
}

// ConditionOnPropertyValue 设置一个 PropertyValueCondition
func (r *Router) ConditionOnPropertyValue(name string, havingValue interface{}) *Router {
	r.cond.OnPropertyValue(name, havingValue)
	return r
}

// ConditionOnBean 设置一个 BeanCondition
func (r *Router) ConditionOnBean(selector interface{}) *Router {
	r.cond.OnBean(selector)
	return r
}

// ConditionOnMissingBean 设置一个 MissingBeanCondition
func (r *Router) ConditionOnMissingBean(selector interface{}) *Router {
	r.cond.OnMissingBean(selector)
	return r
}

// ConditionOnExpression 设置一个 ExpressionCondition
func (r *Router) ConditionOnExpression(expression string) *Router {
	r.cond.OnExpression(expression)
	return r
}

// ConditionOnMatches 设置一个 FunctionCondition
func (r *Router) ConditionOnMatches(fn SpringCore.ConditionFunc) *Router {
	r.cond.OnMatches(fn)
	return r
}

// ConditionOnProfile 设置一个 ProfileCondition
func (r *Router) ConditionOnProfile(profile string) *Router {
	r.cond.OnProfile(profile)
	return r
}

// Request 注册任意 HTTP 方法处理函数
func (r *Router) Request(method uint32, path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	filters = append(r.filters, filters...) // 组合 Router 和 Mapper 的过滤器列表
	return r.mapping.Request(method, r.basePath+path, fn, filters).
		ConditionOn(r.cond).
		OnPorts(r.ports...)
}

// GET 注册 GET 方法处理函数
func (r *Router) GET(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return r.Request(SpringWeb.MethodGet, path, fn, filters...)
}

// HandleGet 注册 GET 方法处理函数
func (r *Router) HandleGet(path string, fn SpringWeb.Handler, filters ...SpringWeb.Filter) *Mapping {
	return r.Request(SpringWeb.MethodGet, path, fn, filters...)
}

// GetMapping 注册 GET 方法处理函数
func (r *Router) GetMapping(path string, fn SpringWeb.HandlerFunc, filters ...SpringWeb.Filter) *Mapping {
	return r.Request(SpringWeb.MethodGet, path, SpringWeb.FUNC(fn), filters...)
}

// GetBinding 注册 GET 方法处理函数
func (r *Router) GetBinding(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return r.Request(SpringWeb.MethodGet, path, SpringWeb.BIND(fn), filters...)
}

// POST 注册 POST 方法处理函数
func (r *Router) POST(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return r.Request(SpringWeb.MethodPost, path, fn, filters...)
}

// HandlePost 注册 POST 方法处理函数
func (r *Router) HandlePost(path string, fn SpringWeb.Handler, filters ...SpringWeb.Filter) *Mapping {
	return r.Request(SpringWeb.MethodPost, path, fn, filters...)
}

// PostMapping 注册 POST 方法处理函数
func (r *Router) PostMapping(path string, fn SpringWeb.HandlerFunc, filters ...SpringWeb.Filter) *Mapping {
	return r.Request(SpringWeb.MethodPost, path, SpringWeb.FUNC(fn), filters...)
}

// PostBinding 注册 POST 方法处理函数
func (r *Router) PostBinding(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return r.Request(SpringWeb.MethodPost, path, SpringWeb.BIND(fn), filters...)
}

// PATCH 注册 PATCH 方法处理函数
func (r *Router) PATCH(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return r.Request(SpringWeb.MethodPatch, path, fn, filters...)
}

// PUT 注册 PUT 方法处理函数
func (r *Router) PUT(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return r.Request(SpringWeb.MethodPut, path, fn, filters...)
}

// DELETE 注册 DELETE 方法处理函数
func (r *Router) DELETE(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return r.Request(SpringWeb.MethodDelete, path, fn, filters...)
}

// HEAD 注册 HEAD 方法处理函数
func (r *Router) HEAD(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return r.Request(SpringWeb.MethodHead, path, fn, filters...)
}

// OPTIONS 注册 OPTIONS 方法处理函数
func (r *Router) OPTIONS(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return r.Request(SpringWeb.MethodOptions, path, fn, filters...)
}

///////////////////// 以下是全局函数 /////////////////////////////

// DefaultWebMapping 默认的 Web 路由映射表
var DefaultWebMapping = NewWebMapping()

// Route 返回和 Mapping 绑定的路由分组
func Route(basePath string, filters ...SpringWeb.Filter) *Router {
	return newRouter(DefaultWebMapping, basePath, filters)
}

// Request 注册任意 HTTP 方法处理函数
func Request(method uint32, path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return DefaultWebMapping.Request(method, path, fn, filters)
}

// GET 注册 GET 方法处理函数
func GET(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return Request(SpringWeb.MethodGet, path, fn, filters...)
}

// HandleGet 注册 GET 方法处理函数
func HandleGet(path string, fn SpringWeb.Handler, filters ...SpringWeb.Filter) *Mapping {
	return Request(SpringWeb.MethodGet, path, fn, filters...)
}

// GetMapping 注册 GET 方法处理函数
func GetMapping(path string, fn SpringWeb.HandlerFunc, filters ...SpringWeb.Filter) *Mapping {
	return Request(SpringWeb.MethodGet, path, SpringWeb.FUNC(fn), filters...)
}

// GetBinding 注册 GET 方法处理函数
func GetBinding(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return Request(SpringWeb.MethodGet, path, SpringWeb.BIND(fn), filters...)
}

// POST 注册 POST 方法处理函数
func POST(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return Request(SpringWeb.MethodPost, path, fn, filters...)
}

// HandlePost 注册 POST 方法处理函数
func HandlePost(path string, fn SpringWeb.Handler, filters ...SpringWeb.Filter) *Mapping {
	return Request(SpringWeb.MethodPost, path, fn, filters...)
}

// PostMapping 注册 POST 方法处理函数
func PostMapping(path string, fn SpringWeb.HandlerFunc, filters ...SpringWeb.Filter) *Mapping {
	return Request(SpringWeb.MethodPost, path, SpringWeb.FUNC(fn), filters...)
}

// PostBinding 注册 POST 方法处理函数
func PostBinding(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return Request(SpringWeb.MethodPost, path, SpringWeb.BIND(fn), filters...)
}

// PUT 注册 PUT 方法处理函数
func PUT(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return Request(SpringWeb.MethodPut, path, fn, filters...)
}

// PATCH 注册 PATCH 方法处理函数
func PATCH(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return Request(SpringWeb.MethodPatch, path, fn, filters...)
}

// DELETE 注册 DELETE 方法处理函数
func DELETE(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return Request(SpringWeb.MethodDelete, path, fn, filters...)
}

// HEAD 注册 HEAD 方法处理函数
func HEAD(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return Request(SpringWeb.MethodHead, path, fn, filters...)
}

// OPTIONS 注册 OPTIONS 方法处理函数
func OPTIONS(path string, fn interface{}, filters ...SpringWeb.Filter) *Mapping {
	return Request(SpringWeb.MethodOptions, path, fn, filters...)
}

///////////////////// Web Filter //////////////////////

// WebFilter 封装一个 SpringWeb.Filter 对象或者一个 Bean 选择器
type WebFilter struct {
	filter SpringWeb.Filter        // Filter 对象
	beanId string                  // Bean 选择器
	cond   *SpringCore.Conditional // 判断条件
}

// Filter 封装一个 SpringWeb.Filter 对象
func Filter(filter SpringWeb.Filter) *WebFilter {
	return &WebFilter{
		filter: filter,
		cond:   SpringCore.NewConditional(),
	}
}

// FilterBean 封装一个 Bean 选择器
func FilterBean(selector interface{}) *WebFilter {

	// TODO 重复代码
	var beanId string
	switch s := selector.(type) {
	case string:
		beanId = s
	default:
		t := reflect.TypeOf(s) // map、slice 等不是指针类型
		if t.Kind() == reflect.Ptr {
			e := t.Elem()
			if e.Kind() == reflect.Interface {
				t = e // 接口类型去掉指针
			}
		}
		beanId = SpringCore.TypeName(t) + ":"
	}

	return &WebFilter{
		beanId: beanId,
		cond:   SpringCore.NewConditional(),
	}
}

func (f *WebFilter) Filter() SpringWeb.Filter {
	return f.filter
}

func (f *WebFilter) FilterBean() string {
	return f.beanId
}

func (f *WebFilter) Invoke(ctx SpringWeb.WebContext, chain SpringWeb.FilterChain) {
	panic(errors.New("shouldn't call this method"))
}

// Or c=a||b
func (f *WebFilter) Or() *WebFilter {
	f.cond.Or()
	return f
}

// And c=a&&b
func (f *WebFilter) And() *WebFilter {
	f.cond.And()
	return f
}

// ConditionOn 设置一个 Condition
func (f *WebFilter) ConditionOn(cond SpringCore.Condition) *WebFilter {
	f.cond.OnCondition(cond)
	return f
}

// ConditionNot 设置一个取反的 Condition
func (f *WebFilter) ConditionNot(cond SpringCore.Condition) *WebFilter {
	f.cond.OnConditionNot(cond)
	return f
}

// ConditionOnProperty 设置一个 PropertyCondition
func (f *WebFilter) ConditionOnProperty(name string) *WebFilter {
	f.cond.OnProperty(name)
	return f
}

// ConditionOnMissingProperty 设置一个 MissingPropertyCondition
func (f *WebFilter) ConditionOnMissingProperty(name string) *WebFilter {
	f.cond.OnMissingProperty(name)
	return f
}

// ConditionOnPropertyValue 设置一个 PropertyValueCondition
func (f *WebFilter) ConditionOnPropertyValue(name string, havingValue interface{}) *WebFilter {
	f.cond.OnPropertyValue(name, havingValue)
	return f
}

// ConditionOnBean 设置一个 BeanCondition
func (f *WebFilter) ConditionOnBean(selector interface{}) *WebFilter {
	f.cond.OnBean(selector)
	return f
}

// ConditionOnMissingBean 设置一个 MissingBeanCondition
func (f *WebFilter) ConditionOnMissingBean(selector interface{}) *WebFilter {
	f.cond.OnMissingBean(selector)
	return f
}

// ConditionOnExpression 设置一个 ExpressionCondition
func (f *WebFilter) ConditionOnExpression(expression string) *WebFilter {
	f.cond.OnExpression(expression)
	return f
}

// ConditionOnMatches 设置一个 FunctionCondition
func (f *WebFilter) ConditionOnMatches(fn SpringCore.ConditionFunc) *WebFilter {
	f.cond.OnMatches(fn)
	return f
}

// ConditionOnProfile 设置一个 ProfileCondition
func (f *WebFilter) ConditionOnProfile(profile string) *WebFilter {
	f.cond.OnProfile(profile)
	return f
}

// Matches 成功返回 true，失败返回 false
func (f *WebFilter) Matches(ctx SpringCore.SpringContext) bool {
	return f.cond.Matches(ctx)
}
