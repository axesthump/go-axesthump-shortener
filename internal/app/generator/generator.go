package generator

import (
	"context"
	"sync"
)

type IDGenerator struct {
	id     int64
	idCh   chan int64
	ctx    context.Context
	cancel context.CancelFunc
}

var idGenerator *IDGenerator
var sy sync.Once

func GetIDGenerator(startID int64) *IDGenerator {
	sy.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		idGenerator = &IDGenerator{
			id:     startID,
			idCh:   make(chan int64),
			ctx:    ctx,
			cancel: cancel,
		}
		go idGenerator.start()
	})
	return idGenerator
}

func (g *IDGenerator) GetID() int64 {
	return <-g.idCh
}

func (g *IDGenerator) start() {
	for {
		select {
		case <-g.ctx.Done():
			return
		default:
			g.idCh <- g.id
			g.id++
		}
	}
}

func (g *IDGenerator) Cancel() {
	g.cancel()
}
