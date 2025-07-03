package main

type Dependency[T any] interface {
	Update(T)
}

type Signal[T any] struct {
	value T
	deps  []Dependency[T]
}

func (s *Signal[T]) Value() T {
	return s.value
}

func (s *Signal[T]) Set(val T) {
	s.value = val
	for _, dep := range s.deps {
		dep.Update(val)
	}
}

func (s *Signal[T]) Update(val T) {
	s.Set(val)
}

func (s *Signal[T]) AddDependency(dep Dependency[T]) {
	s.deps = append(s.deps, dep)
}

func (s *Signal[T]) Effect(f func()) {
	s.deps = append(s.deps, EffectFunc[T](f))
}

type EffectFunc[T any] func()

func (e EffectFunc[T]) Update(_ T) {
	e()
}
