//go:build linux

package drivers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGPIOCDevFactory(t *testing.T) {
	t.Parallel()

	f := NewGPIOCDevFactory()
	require.NotNil(t, f)
	assert.IsType(t, &GPIOCDevFactory{}, f)
}
