package testify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type compStruct struct {
	b bool
	i int
	f float32
	s string
}

func TestA(t *testing.T) {
	assert.Equal(t, true, false)
	assert.Nil(t, true)
}

func TestB(t *testing.T) {
	t.Run("testc", func(t *testing.T) {
		a := compStruct{
			b: true,
			i: 20,
			f: 54,
			s: "first",
		}
		b := compStruct{
			b: false,
			i: 24,
			f: 54,
			s: "second",
		}
		assert.Equal(t, a, b)
	})
	t.Run("testd", func(t *testing.T) {
	})
	t.Run("teste", func(t *testing.T) {
	})
}
