package clickhouse

import (
	"errors"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/valyala/fastrand"
)

var (
	ErrNoServerAvailable = errors.New("balance select error: server is not available")
)

type Server struct {
	DB     *sqlx.DB
	Name   string
	Weight int
	// 主机是否在线.
	// Online bool.
}
type LoadBalance interface {
	// 选择一个后端Server.
	// 参数remove是需要排除选择的后端Server.
	Select(remove []*sqlx.DB) (*Server, error)
	// 更新可用Server列表.
	UpdateServers(servers []*Server)
	Close()
}

type LoadBalanceRandom struct {
	servers          []*Server
	notActiveServers []*Server
	mutex            sync.Mutex
}

func NewLoadBalanceRandom(servers []*Server) *LoadBalanceRandom {
	newBalance := &LoadBalanceRandom{}
	newBalance.UpdateServers(servers)
	return newBalance
}
func (l *LoadBalanceRandom) Close() {
	for _, v := range l.servers {
		v.DB.Close()
	}
}

// 系统运行过程中，后端可用Server会更新.
func (l *LoadBalanceRandom) UpdateServers(servers []*Server) {
	newServers := make([]*Server, 0)
	newServers = append(newServers, servers...)
	l.servers = newServers
	l.notActiveServers = make([]*Server, 0)
	go l.daemon()
}

func (l *LoadBalanceRandom) daemon() {
	for {
		l.mutex.Lock()
		tempNotActiveServers := make([]*Server, 0)
		tempActiveServers := make([]*Server, 0)
		notActiveServers := l.notActiveServers
		for _, notActiveServer := range notActiveServers {
			if notActiveServer.DB.Ping() != nil {
				tempNotActiveServers = append(tempNotActiveServers, notActiveServer)
			} else {
				tempActiveServers = append(tempActiveServers, notActiveServer)
			}
		}
		l.notActiveServers = tempNotActiveServers
		l.servers = append(l.servers, tempActiveServers...)
		l.mutex.Unlock()
		time.Sleep(time.Second * 5)
	}
}

// 选择一个后端Server.
func (l *LoadBalanceRandom) Select(remove []*sqlx.DB) (*Server, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	curServer := l.servers
	if len(curServer) == 0 {
		return nil, ErrNoServerAvailable
	}

	if len(remove) == 0 {
		return l.selectServer(curServer, []*Server{})
	}
	tmpServers := make([]*Server, 0)
	removeServers := make([]*Server, 0)
	for _, s := range curServer {
		isFind := false
		for _, v := range remove {
			if s.DB == v {
				isFind = true
				removeServers = append(removeServers, s)
				break
			}
		}
		if !isFind {
			tmpServers = append(tmpServers, s)
		}
	}
	if len(tmpServers) == 0 {
		return nil, ErrNoServerAvailable
	}
	// selected := fastrand.Uint32n(uint32(len(tmpServers)))
	// return tmpServers[selected]
	return l.selectServer(tmpServers, removeServers)
}

func (l *LoadBalanceRandom) selectServer(activeServers []*Server, removeServers []*Server) (*Server, error) {
	curActiveServer := activeServers
	for i := 0; i < 5; i++ {
		serverLen := uint32(len(curActiveServer))
		if serverLen == 0 {
			return nil, ErrNoServerAvailable
		}
		id := fastrand.Uint32n(serverLen)
		activeServers := make([]*Server, 0)
		notActiveServers := make([]*Server, 0)
		var resultServer *Server
		var isSelected bool
		for i, server := range curActiveServer {
			if server.DB.Ping() != nil {
				notActiveServers = append(notActiveServers, server)
			} else {
				activeServers = append(activeServers, server)
				if id == uint32(i) {
					resultServer = server
					isSelected = true
				}
			}
		}
		if len(removeServers) != 0 && i == 0 {
			for _, removeServer := range removeServers {
				if removeServer.DB.Ping() == nil {
					activeServers = append(activeServers, removeServer)
				} else {
					notActiveServers = append(notActiveServers, removeServer)
				}
			}
		}
		l.servers = activeServers
		l.notActiveServers = append(l.notActiveServers, notActiveServers...)
		curActiveServer = l.servers
		if isSelected {
			return resultServer, nil
		}
	}
	return nil, ErrNoServerAvailable
}

func (l *LoadBalanceRandom) String() string {
	return "Random"
}
