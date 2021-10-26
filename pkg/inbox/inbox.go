package inbox

import (
	"context"
	"runtime"
	"time"
)

type Header struct {
	Key   string
	Value string
}

func NewHeader(key, value string) Header {
	return Header{key, value}
}

type MessageCtx struct {
	Headers []Header
	Offset  Offseter
	Message interface{}
}

type inbox struct {
	Buffer       []MessageCtx
	msgCh        chan MessageCtx
	ticker       *time.Ticker
	recivers     map[string]MsgReciver
	shutdowCh    chan struct{}
	inboxManager InboxManager

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
		shutdowCh:   make(chan struct{}),
		ticker:      time.NewTicker(10),
		expiredTime: defaultExpiredTime,
		msgCh:       make(chan MessageCtx, nonBlockNum),
		Buffer:      make([]MessageCtx, capcity),
	}
}

func (ib *inbox) OnMessage(msg MessageCtx) {
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
				msg := ib.Buffer[(blockIdx+n)%ib.capcity]
				reciverId := msg.Headers[MsgReciverId]
				if reciver, exists := ib.recivers[reciverId]; exists {
					if MsgReciverStatusInactive == reciver.Status() {
						log.Infof("inactive reciver, evicted reciver (%s).", reciverId)
						delete(ib.recivers, reciverId)
					} else {
						_, err := reciver.OnMessage(msg)
						if nil != err {
							// Entity 负载达到上限，跟不上，那么我们现在实现的策略为 阻塞等待.
							runtime.Gosched()
							break
						}
					}
				}

				log.Infof("handle msg: %v.", msg)

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
	var (
		headIdx = ib.headIdx
		offset  = NewOffseter()
	)

	if ib.size != 0 {
		num := ib.size - ib.blockNum
		for index := 0; index < num; index++ {
			msg := ib.Buffer[headIdx]
			if !msg.Offset.Status() {
				break
			} else if !msg.Offset.AutoCommit() {
				offset = msg.Offset
				headIdx++
				continue
			}
		}

		if headIdx == ib.headIdx {
			return false
		}

		if err := offset.Commit(); nil != err {
			log.Errorf("commit failed, %s.", err.Error())
			return false
		}

		ib.headIdx = headIdx
	}

	return true
}

func (ib *inbox) evictedHead() {
	// 先直接跳过队头阻塞的消息.
	if ib.size == 0 {
		return
	}

	headIdx := ib.headIdx
	msg0 := ib.Buffer[headIdx]
	reciverId = msg0.Headers[MsgReciverId]

	for i := 0; i < ib.size; i++ {
		msg := ib.Buffer[(headIdx)%ib.capcity]
		if msg.Headers[MsgReciverId] != reciverId {
			break
		}
		headIdx++
	}

	ib.size
	ib.headIdx
}
