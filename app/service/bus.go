package service

import (
	"context"

	"github.com/lancer-kit/uwe/v2"
)

type Message struct {
	UID    int64
	Target uwe.WorkerName
	Sender uwe.WorkerName
	Data   interface{}
}

type EventBus interface {
	context.Context
	SendMessage(target uwe.WorkerName, data interface{}) error
	MessageBus() <-chan *Message
}

type eventBus struct {
	context.Context

	name uwe.WorkerName

	// in is channel for incoming messages for a worker
	in chan *Message
	// out is channel for outgoing messages from a worker
	out chan<- *Message
}

func NewEventBus(name uwe.WorkerName, ctx context.Context, toWorker, fromWorker chan *Message) EventBus {
	return &eventBus{
		Context: ctx,
		name:    name,
		in:      toWorker,
		out:     fromWorker,
	}

}

func (wc *eventBus) SendMessage(target uwe.WorkerName, data interface{}) error {
	wc.out <- &Message{
		UID:    0,
		Target: target,
		Sender: wc.name,
		Data:   data,
	}
	return nil
}

func (wc *eventBus) MessageBus() <-chan *Message {
	return wc.in
}

type EventHub struct {
	ctx        context.Context
	cancelFunc context.CancelFunc

	defaultChanLen  int
	workersHub      map[uwe.WorkerName]chan<- *Message
	workersMessages chan *Message
}

func NewEventHub(defaultChanLen int) *EventHub {
	if defaultChanLen < 1 {
		defaultChanLen = 1
	}
	ctx, cancel := context.WithCancel(context.Background())

	return &EventHub{
		ctx:             ctx,
		cancelFunc:      cancel,
		workersHub:      map[uwe.WorkerName]chan<- *Message{},
		workersMessages: make(chan *Message, defaultChanLen)}
}

func (hub *EventHub) AddWorker(name uwe.WorkerName) EventBus {
	workerDirectChan := make(chan *Message, hub.defaultChanLen)
	hub.workersHub[name] = workerDirectChan
	return NewEventBus(name, hub.ctx, workerDirectChan, hub.workersMessages)
}

func (hub *EventHub) Init() error {
	return nil
}

func (hub *EventHub) Run(ctx uwe.Context) error {
	for {
		select {
		case m := <-hub.workersMessages:
			if m == nil {
				continue
			}

			switch m.Target {
			// todo: review this case
			case "connect", "subscriber":
				_, ok := hub.workersHub[m.Sender]
				if !ok {
					hub.workersHub[m.Sender] = make(chan *Message, len(hub.workersHub))
				}

			case "*", "broadcast":
				for to := range hub.workersHub {
					hub.workersHub[to] <- m
				}
			default:
				if _, ok := hub.workersHub[m.Target]; ok {
					hub.workersHub[m.Target] <- m
				}
			}

		case <-ctx.Done():
			hub.cancelFunc()
			return nil
		}
	}
}
