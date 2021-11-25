package volumeerror

//ErrorKind : for Specifying the type of error
type ErrorKind string

// Type of volume error
const (
	ErrorKindBestFitFailed ErrorKind = "BestFitFailed"
)

//Error : Struct to Encapsulate the error and with ErrorKind
type Error struct {
	Kind ErrorKind
	Err  error
}

func (e *Error) Error() string {
	return e.Err.Error()
}
