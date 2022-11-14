package generator

import (
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
