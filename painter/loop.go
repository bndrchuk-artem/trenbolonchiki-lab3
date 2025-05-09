package painter

import (
	"image"
	"sync"

	"golang.org/x/exp/shiny/screen"
)


type Receiver interface {
	Update(t screen.Texture)
}

type Loop struct {
	Receiver Receiver

	next screen.Texture
	prev screen.Texture

	stopReq bool
	stopped chan struct{}

	MsgQueue messageQueue
}

var size = image.Pt(800, 800)

func (l *Loop) Start(s screen.Screen) {
	l.next, _ = s.NewTexture(size)
	l.prev, _ = s.NewTexture(size)

	l.MsgQueue = messageQueue{}
	go l.eventProcess()
}

func (l *Loop) eventProcess() {
	for {
		if op := l.MsgQueue.Pull(); op != nil {
			if update := op.Do(l.next); update {
				l.Receiver.Update(l.next)
				l.next, l.prev = l.prev, l.next
			}
		}
	}
}

func (l *Loop) Post(op Operation) {
	if op != nil {
		l.MsgQueue.Push(op)
	}
}

func (l *Loop) StopAndWait() {
	l.Post(OperationFunc(func(screen.Texture) {
		l.stopReq = true
	}))
	<-l.stopped
}

type messageQueue struct {
	Queue   []Operation
	mu      sync.Mutex
	blocked chan struct{}
}

func (MsgQueue *messageQueue) Push(op Operation) {
	MsgQueue.mu.Lock()
	defer MsgQueue.mu.Unlock()
	MsgQueue.Queue = append(MsgQueue.Queue, op)
	if MsgQueue.blocked != nil {
		close(MsgQueue.blocked)
		MsgQueue.blocked = nil
	}
}

func (MsgQueue *messageQueue) Pull() Operation {
	MsgQueue.mu.Lock()
	defer MsgQueue.mu.Unlock()
	for len(MsgQueue.Queue) == 0 {
		MsgQueue.blocked = make(chan struct{})
		MsgQueue.mu.Unlock()
		<-MsgQueue.blocked
		MsgQueue.mu.Lock()
	}
	op := MsgQueue.Queue[0]
	MsgQueue.Queue[0] = nil
	MsgQueue.Queue = MsgQueue.Queue[1:]
	return op
}
