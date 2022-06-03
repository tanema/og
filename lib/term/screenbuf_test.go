package term

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuncMap(t *testing.T) {
	str := parseAnsiString("hello")
	str.add(fg + white)
	assert.Equal(t, esc+fg+white+"mhello"+esc+"m", str.String())
	str.add(bold)
	assert.Equal(t, esc+fg+white+";"+bold+"mhello"+esc+"m", str.String())
	str.replace(fg, brfg)
	assert.Equal(t, esc+brfg+white+";"+bold+"mhello"+esc+"m", str.String())
	str = parseAnsiString("hello" + str.String())
	str.add(bg + green)
	assert.Equal(t, esc+bg+green+"mhello"+esc+brfg+white+";"+bold+"mhello"+esc+"m"+esc+"m", str.String())
}
