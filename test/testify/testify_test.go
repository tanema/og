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
	d *rune
}

func TestA(t *testing.T) {
	assert.Equal(t, true, false)
	assert.Nil(t, true)
	assert.Equal(t, true, 24, "These should be the wrong type")
}

func TestB(t *testing.T) {
	t.Run("testc", func(t *testing.T) {
		r := 'a'
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
			d: &r,
		}
		assert.Equal(t, a, b, "this is in test c")
	})
	t.Run("testd", func(t *testing.T) {
		assert.Equal(
			t,
			"this is a very long string this is a very long string this is a very long string",
			`this is a very long string
		this is a very long string
		this is a very long string`)
	})
	t.Run("teste", func(t *testing.T) {
		t.Skip()
	})
}
