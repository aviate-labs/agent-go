package cmotoko

// Result is a generic type that represents either success (Ok) or failure (Err).
// It is used as the return type of functions which may fail.
type Result[T any, E any] struct {
	Ok  *T `ic:"ok,variant"`
	Err *E `ic:"err,variant"`
}
