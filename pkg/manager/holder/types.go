package holder

import "context"

type Status string

func (s Status) String() string {
	return string(s)
}

const (
	StatusOK       Status = "OK"
	StatusError    Status = "Error"
	StatusCanceled Status = "Canceled"
)

type Holder interface {
	Wait(ctx context.Context, id string) Response
	OnRespond(*Response)
}

type Response struct {
	ID       string
	Status   Status
	ErrCode  string
	Metadata map[string]string
	Data     []byte
}
