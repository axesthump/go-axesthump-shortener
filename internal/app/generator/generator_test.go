package generator

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewIDGenerator(t *testing.T) {
	g := NewIDGenerator(0)
	defer g.Cancel()
	assert.NotNil(t, g.idCh)
	assert.Equal(t, int64(0), g.GetID())
	g.Cancel()
}

func TestIDGenerator_GetID(t *testing.T) {
	g := NewIDGenerator(0)
	defer g.Cancel()

	id := g.GetID()
	assert.Equal(t, int64(0), id)
	id = g.GetID()
	assert.Equal(t, int64(1), id)

}

func TestIDGenerator_IsCreatedID(t *testing.T) {
	g := NewIDGenerator(0)
	defer g.Cancel()

	id := g.GetID()
	id2 := g.GetID()
	fmt.Println(id, id2)
	assert.True(t, g.IsCreatedID(0))
	assert.True(t, g.IsCreatedID(1))
	assert.False(t, g.IsCreatedID(2))

}
