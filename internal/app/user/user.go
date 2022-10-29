package user

import "sync"

type IDGenerator struct {
	nextID uint32
	mx     sync.RWMutex
}

func (g *IDGenerator) GetNewUserID() uint32 {
	g.mx.Lock()
	defer g.mx.Unlock()
	res := g.nextID
	g.nextID++
	return res
}

func (g *IDGenerator) IsCreatedUser(id uint32) bool {
	g.mx.RLock()
	defer g.mx.RUnlock()
	res := id < g.nextID
	return res
}
