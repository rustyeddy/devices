package drivers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, Bias("default"), BiasDefault)
	assert.Equal(t, Bias("pullup"), BiasPullUp)
	assert.Equal(t, Bias("pulldown"), BiasPullDown)

	assert.Equal(t, Edge("none"), EdgeNone)
	assert.Equal(t, Edge("rising"), EdgeRising)
	assert.Equal(t, Edge("falling"), EdgeFalling)
	assert.Equal(t, Edge("both"), EdgeBoth)
}
