package placement

import "github.com/tkeel-io/core/pkg/repository/dao"

type Placement interface {
	Select(string) dao.Queue
	AppendQueue(dao.Queue)
	RemoveQueue(dao.Queue)
}

func Global() Placement {
	return globalPlacement
}
