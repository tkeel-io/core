package inbox

import (
	"context"
	"runtime"
	"time"
)

type IbElem struct {
	Index   int
	Offset  Offseter
	Message interface{}
}

type inbox struct {
	Elems     []IbElem
	msgCh     chan IbElem
	ticker    *time.Ticker
	msgHandle MessageHandler
	shutdowCh chan struct{}

	size        int
	capcity     int
	headIdx     int
	blockNum    int
	lastCommit  int64
	expiredTime int

	ctx    context.Context
	cancel context.CancelFunc
}

// NewInbox returns a inbox instance.
func NewInbox(ctx context.Context, capcity, nonBlockNum int, msgHandle MessageHandler) Inbox {
	ctx, cancel := context.WithCancel(ctx)

	if nonBlockNum < 10 {
		nonBlockNum = defaultNonBlockNum
	}

	return &inbox{
		ctx:         ctx,
		cancel:      cancel,
		size:        0,
		capcity:     capcity,
		msgHandle:   msgHandle,
		shutdowCh:   make(chan struct{}),
		ticker:      time.NewTicker(10),
		expiredTime: defaultExpiredTime,
		msgCh:       make(chan IbElem, nonBlockNum),
		Elems:       make([]IbElem, capcity),
	}
}

func (ib *inbox) OnMessage(msg IbElem) {
	ib.msgCh <- msg
}

func (ib *inbox) Start() { // nolint
	log.Info("inbox start...")

	for {
		select {
		case <-ib.ctx.Done():
			ib.cancel()
			ib.Stop()
		case <-ib.shutdowCh:
			ib.cancel()
			ib.Stop()
		default:
			// recive msg from msgCh.
			idelNum := ib.capcity - ib.size
			for n := 0; n < idelNum; n++ {
				select {
				case msg := <-ib.msgCh:
					ib.Elems[(ib.headIdx+ib.size)%ib.capcity] = msg
					ib.size++
				default:
					break
				}
			}

			// handle msg.
			blockNum := ib.blockNum
			blockIdx := (ib.headIdx + ib.size - ib.blockNum) % ib.capcity
			for n := 0; n < blockNum; n++ {
				_, err := ib.msgHandle(ib.Elems[(blockIdx+n)%ib.capcity])
				if nil != err {
					// Entity 负载达到上限，跟不上，那么我们现在实现的策略为 阻塞等待.
					runtime.Gosched()
					break
				}

				ib.blockNum--
			}

			// commit.
			if ib.commit() {
				ib.lastCommit = time.Now().UnixNano() / 1e6
			} else {
				if time.Now().UnixNano()/1e6-ib.lastCommit > int64(ib.expiredTime) {
					ib.evictedHead()
					ib.lastCommit = time.Now().UnixNano() / 1e6
				}
				runtime.Gosched()
			}
		}
	}
}

func (ib *inbox) Stop() {}

func (ib *inbox) commit() bool {
	return true
}

func (ib *inbox) evictedHead() {

}
