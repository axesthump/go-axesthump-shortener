// Package generator define IDGenerator to generate a unique id.
package generator

import (
	"context"
)

// IDGenerator contains data for generate a unique id.
type IDGenerator struct {
	id     int64
	idCh   chan int64
	ctx    context.Context
	cancel context.CancelFunc
}

// NewIDGenerator returns new IDGenerator where unique id start with startID.
func NewIDGenerator(startID int64) *IDGenerator {
	ctx, cancel := context.WithCancel(context.Background())
	idGenerator := &IDGenerator{
		id:     startID,
		idCh:   make(chan int64),
		ctx:    ctx,
		cancel: cancel,
	}
	go idGenerator.start()
	return idGenerator
}

// GetID returns new unique id.
func (g *IDGenerator) GetID() int64 {
	return <-g.idCh
}

// Cancel stopping process id generation.
func (g *IDGenerator) Cancel() {
	g.cancel()
}

// IsCreatedID checks id is created.
func (g *IDGenerator) IsCreatedID(id uint32) bool {
	return int64(id) < g.id
}

// start process generation unique id.
func (g *IDGenerator) start() {
	for {
		select {
		case <-g.ctx.Done():
			close(g.idCh)
			return
		default:
			id := g.id
			g.id++
			g.idCh <- id
		}
	}
}
