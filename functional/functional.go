package functional

func Cast[T any](a any) T {
	return a.(T)
}

func CastArray[T any](a []any) []T {
	return Map(Cast[T], a)
}

func Map[T, U any](f func(T) U, a []T) []U {
	res := make([]U, len(a))
	for i, v := range a {
		res[i] = f(v)
	}
	return res
}

func Filter[T any](pred func(T) bool, a []T) []T {
	res := make([]T, 0, len(a))
	for _, v := range a {
		if pred(v) {
			_ = append(res, v)
		}
	}
	return res
}

func Partial[T any](fn func(T, ...any) any, val T) func(...any) any {
	return func(args ...any) any {
		return fn(val, args...)
	}
}

func Foldl[T, U any](f func(U, T) U, z U, a []T) U {
	r := z
	for _, v := range a {
		r = f(r, v)
	}
	return r
}

func Foldl1[T any](f func(T, T) T, a []T) T {
	return Foldl(f, a[0], a[1:])
}

func Foldr[T, U any](f func(T, U) U, z U, a []T) U {
	r := z
	for i := len(a) - 1; i >= 0; i-- {
		r = f(a[i], r)
	}
	return r
}

func Foldr1[T any](f func(T, T) T, a []T) T {
	n := len(a) - 1
	return Foldl(f, a[n], a[:n])
}

// type Monad [A any] interface {
// //	Bind_(b Monad[B])Monad[B]       // >>    m = \_ -> b
// }

// func MonadBind[A, B any](a Monad[A], m func(A)Monad[B])Monad[B] { // >>=
// 	return m(a)
// }
