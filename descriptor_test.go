package devices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type describedFixture struct{}

func (describedFixture) Descriptor() Descriptor {
	return Descriptor{
		Name:      "dev",
		Kind:      "sensor",
		ValueType: "float64",
		Access:    ReadOnly,
		Unit:      "C",
		Min:       floatPtr(-10),
		Max:       floatPtr(100),
		Tags:      []string{"t1", "t2"},
		Attributes: map[string]string{
			"key": "value",
		},
	}
}

func TestAccessModeConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, AccessMode("ro"), ReadOnly)
	assert.Equal(t, AccessMode("wo"), WriteOnly)
	assert.Equal(t, AccessMode("rw"), ReadWrite)
}

func TestDescriptorFields(t *testing.T) {
	t.Parallel()

	var d Described = describedFixture{}
	desc := d.Descriptor()

	assert.Equal(t, "dev", desc.Name)
	assert.Equal(t, "sensor", desc.Kind)
	assert.Equal(t, "float64", desc.ValueType)
	assert.Equal(t, ReadOnly, desc.Access)
	assert.Equal(t, "C", desc.Unit)
	assert.Equal(t, -10.0, *desc.Min)
	assert.Equal(t, 100.0, *desc.Max)
	assert.Equal(t, []string{"t1", "t2"}, desc.Tags)
	assert.Equal(t, map[string]string{"key": "value"}, desc.Attributes)
}

func floatPtr(v float64) *float64 {
	return &v
}
