package internal

type CtxKey struct {
	Key string
}

func Unique[T comparable](in []T) []T {
	seen := make(map[T]struct{})
	out := make([]T, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}
