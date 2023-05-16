package expandable

import (
	"fmt"
	"reflect"
	"unsafe"
)

var nextBaseIndex = 0

type Schema struct {
	name  string
	types []any
}

func NewSchema(name string) *Schema {
	return &Schema{
		name: name,
	}
}

type Any[B any] interface {
	String() string
	newThis() *B
	asThis(base *B) any
}

type Base[B any] struct {
	index  int
	schema *Schema
	name   string
	typ    reflect.Type
	new    func() *B
	full   Any[B]
}

func NewBase[B any](schema *Schema, name string) *Base[B] {
	t := &Base[B]{
		index:  nextBaseIndex,
		schema: schema,
		name:   name,
		typ:    reflect.TypeOf((*B)(nil)).Elem(),
	}
	t.full = t
	nextBaseIndex++
	return t
}
func (t *Base[B]) WithNew(f func() *B) *Base[B] {
	t.new = f
	return t
}

func (t *Base[B]) String() string {
	return t.schema.name + "." + t.name
}

func (t *Base[B]) New() *B {
	return t.full.newThis()
}

func (t *Base[B]) Full(base *B) any {
	return t.full.asThis(base)
}

func (t *Base[B]) newThis() *B {
	if t.new != nil {
		return t.new()
	} else {
		return reflect.New(t.typ).Interface().(*B)
	}
}

func (t *Base[B]) asThis(base *B) any {
	return base
}

func (t *Base[B]) addDerived(d Any[B]) {
	if t.full != t {
		panic(fmt.Errorf("trying to derive another %s when %s has already been derived", d.String(), t.full.String()))
	}
	t.full = d
}

type Derived[T, B any] struct {
	schema *Schema
	typ    reflect.Type
	base   *Base[B]
	new    func() *T
}

func Derive[T, B any](schema *Schema, base *Base[B]) *Derived[T, B] {
	t := &Derived[T, B]{
		schema: schema,
		typ:    reflect.TypeOf((*T)(nil)).Elem(),
		base:   base,
	}
	base.addDerived(t)
	return t
}
func (t *Derived[T, B]) WithNew(f func() *T) *Derived[T, B] {
	t.new = f
	return t
}

func (t *Derived[T, B]) String() string {
	return t.schema.name + "." + t.base.name
}

func (t *Derived[T, B]) newThis() *B {
	if t.new != nil {
		return t.Base(t.new())
	} else {
		return t.Base(reflect.New(t.typ).Interface().(*T))
	}
}
func (t *Derived[T, B]) From(base *B) *T {
	return (*T)(unsafe.Pointer(base))
}
func (t *Derived[T, B]) Base(drvd *T) *B {
	return (*B)(unsafe.Pointer(drvd))
}
func (t *Derived[T, B]) asThis(base *B) any {
	return t.From(base)
}
func (t *Derived[T, B]) Wrap(f func(*T)) func(*B) {
	return func(base *B) {
		f(t.From(base))
	}
}
func (t *Derived[T, B]) WrapAE(f func(*T) (any, error)) func(*B) (any, error) {
	return func(base *B) (any, error) {
		return f(t.From(base))
	}
}

func Wrap2[T1, T2, B1, B2 any](f func(*T1, *T2), d1 *Derived[T1, B1], d2 *Derived[T2, B2]) func(*B1, *B2) {
	return func(v1 *B1, v2 *B2) {
		f(d1.From(v1), d2.From(v2))
	}
}

func Wrap21[T1, T2, B1 any](f func(*T1, *T2), d1 *Derived[T1, B1]) func(*B1, *T2) {
	return func(v1 *B1, v2 *T2) {
		f(d1.From(v1), v2)
	}
}

func Wrap21A[T1, T2, B1 any](f func(*T1, *T2) any, d1 *Derived[T1, B1]) func(*B1, *T2) any {
	return func(v1 *B1, v2 *T2) any {
		return f(d1.From(v1), v2)
	}
}

func Wrap2E[T1, T2, B1, B2 any](f func(*T1, *T2) error, d1 *Derived[T1, B1], d2 *Derived[T2, B2]) func(*B1, *B2) error {
	return func(v1 *B1, v2 *B2) error {
		return f(d1.From(v1), d2.From(v2))
	}
}

func Wrap1RE[T1, R1, B1 any](f func(*T1) (R1, error), d1 *Derived[T1, B1]) func(*B1) (R1, error) {
	return func(v1 *B1) (R1, error) {
		return f(d1.From(v1))
	}
}
