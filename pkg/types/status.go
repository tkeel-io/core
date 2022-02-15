package types

type Status string

func (s Status) String() string {
	return string(s)
}

const (
	StatusOK       Status = "OK"
	StatusError    Status = "Error"
	StatusCanceled Status = "Canceled"
)
