package volumeerror

type ErrorKind string

const (
	ErrorKindBestFitFailed ErrorKind = "BestFitFailed"
)

type Error struct {
	Kind ErrorKind
	Err  error
}

func (e *Error) Error() string {
	return e.Err.Error()
}
