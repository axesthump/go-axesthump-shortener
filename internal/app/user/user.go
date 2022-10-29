package user

import "sync"

type IDGenerator struct {
	nextId uint32
	mx     sync.RWMutex
}

func (g *IDGenerator) GetNewUserId() uint32 {
	g.mx.Lock()
	defer g.mx.Unlock()
	res := g.nextId
	g.nextId++
	return res
}

func (g *IDGenerator) IsCreatedUser(id uint32) bool {
	g.mx.RLock()
	defer g.mx.RUnlock()
	res := id < g.nextId
	return res
}
