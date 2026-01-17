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

func TestDescriptorZeroValue(t *testing.T) {
	t.Parallel()

	var d Descriptor

	assert.Equal(t, "", d.Name)
	assert.Equal(t, "", d.Kind)
	assert.Equal(t, "", d.ValueType)
	assert.Equal(t, AccessMode(""), d.Access)
	assert.Equal(t, "", d.Unit)
	assert.Nil(t, d.Min)
	assert.Nil(t, d.Max)
	assert.Nil(t, d.Tags)
	assert.Nil(t, d.Attributes)
}

func TestDescriptorWithNilMinMax(t *testing.T) {
	t.Parallel()

	d := Descriptor{
		Name:      "sensor",
		Kind:      "temperature",
		ValueType: "float64",
		Access:    ReadOnly,
		Unit:      "C",
		Min:       nil,
		Max:       nil,
	}

	assert.Nil(t, d.Min)
	assert.Nil(t, d.Max)
}

func TestDescriptorWithEmptyCollections(t *testing.T) {
	t.Parallel()

	d := Descriptor{
		Name:       "device",
		Tags:       []string{},
		Attributes: map[string]string{},
	}

	assert.NotNil(t, d.Tags)
	assert.Len(t, d.Tags, 0)
	assert.NotNil(t, d.Attributes)
	assert.Len(t, d.Attributes, 0)
}

func TestDescriptorMinMaxEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		min  *float64
		max  *float64
	}{
		{"same values", floatPtr(10.0), floatPtr(10.0)},
		{"negative range", floatPtr(-100.0), floatPtr(-10.0)},
		{"zero min", floatPtr(0.0), floatPtr(100.0)},
		{"zero max", floatPtr(-100.0), floatPtr(0.0)},
		{"both zero", floatPtr(0.0), floatPtr(0.0)},
		{"large values", floatPtr(-1e9), floatPtr(1e9)},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := Descriptor{
				Name: "test",
				Min:  tt.min,
				Max:  tt.max,
			}

			assert.Equal(t, tt.min, d.Min)
			assert.Equal(t, tt.max, d.Max)
			if tt.min != nil && tt.max != nil {
				assert.Equal(t, *tt.min, *d.Min)
				assert.Equal(t, *tt.max, *d.Max)
			}
		})
	}
}

func TestDescriptorAllAccessModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode AccessMode
		desc string
	}{
		{ReadOnly, "read-only device"},
		{WriteOnly, "write-only device"},
		{ReadWrite, "read-write device"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(string(tt.mode), func(t *testing.T) {
			t.Parallel()

			d := Descriptor{
				Name:   tt.desc,
				Access: tt.mode,
			}

			assert.Equal(t, tt.mode, d.Access)
		})
	}
}

func TestDescriptorWithMultipleTags(t *testing.T) {
	t.Parallel()

	tags := []string{"outdoor", "wireless", "battery-powered", "sensor"}
	d := Descriptor{
		Name: "outdoor-sensor",
		Tags: tags,
	}

	assert.Equal(t, tags, d.Tags)
	assert.Len(t, d.Tags, 4)
	assert.Contains(t, d.Tags, "outdoor")
	assert.Contains(t, d.Tags, "sensor")
}

func TestDescriptorWithComplexAttributes(t *testing.T) {
	t.Parallel()

	attrs := map[string]string{
		"gpio":      "17",
		"i2c":       "0x76",
		"protocol":  "modbus",
		"baudrate":  "9600",
		"location":  "room-101",
		"serial_no": "ABC123XYZ",
	}

	d := Descriptor{
		Name:       "complex-device",
		Attributes: attrs,
	}

	assert.Equal(t, attrs, d.Attributes)
	assert.Len(t, d.Attributes, 6)
	assert.Equal(t, "17", d.Attributes["gpio"])
	assert.Equal(t, "0x76", d.Attributes["i2c"])
}

func TestDescriptorAllValueTypes(t *testing.T) {
	t.Parallel()

	types := []string{
		"bool",
		"int",
		"int64",
		"float32",
		"float64",
		"string",
		"[]byte",
		"struct",
		"interface{}",
	}

	for _, vt := range types {
		vt := vt
		t.Run(vt, func(t *testing.T) {
			t.Parallel()

			d := Descriptor{
				Name:      "device",
				ValueType: vt,
			}

			assert.Equal(t, vt, d.ValueType)
		})
	}
}

func TestDescriptorCommonUnits(t *testing.T) {
	t.Parallel()

	units := []string{
		"C", "F", "K", // temperature
		"%", "ppm", // percentage, parts per million
		"m", "cm", "mm", // distance
		"kg", "g", "mg", // mass
		"V", "A", "W", // electrical
		"rpm", "Hz", // frequency
		"Pa", "bar", "psi", // pressure
		"m/s", "km/h", "mph", // velocity
	}

	for _, unit := range units {
		unit := unit
		t.Run(unit, func(t *testing.T) {
			t.Parallel()

			d := Descriptor{
				Name: "sensor",
				Unit: unit,
			}

			assert.Equal(t, unit, d.Unit)
		})
	}
}

func TestDescriptorKindExamples(t *testing.T) {
	t.Parallel()

	kinds := []string{
		"relay", "button", "led",
		"temperature", "humidity", "pressure",
		"gps", "accelerometer", "gyroscope",
		"camera", "microphone", "speaker",
		"motor", "servo", "stepper",
	}

	for _, kind := range kinds {
		kind := kind
		t.Run(kind, func(t *testing.T) {
			t.Parallel()

			d := Descriptor{
				Name: kind + "-device",
				Kind: kind,
			}

			assert.Equal(t, kind, d.Kind)
		})
	}
}

func TestDescribedInterfaceImplementation(t *testing.T) {
	t.Parallel()

	var _ Described = describedFixture{}

	fixture := describedFixture{}
	desc := fixture.Descriptor()

	assert.Equal(t, "dev", desc.Name)
	assert.Equal(t, "sensor", desc.Kind)
}

func TestDescriptorImmutability(t *testing.T) {
	t.Parallel()

	// Verify that returning a Descriptor by value (not pointer)
	// means modifications don't affect the original
	fixture := describedFixture{}
	desc1 := fixture.Descriptor()
	desc2 := fixture.Descriptor()

	// Modify desc1
	desc1.Name = "modified"
	desc1.Tags = append(desc1.Tags, "new-tag")

	// desc2 should be unaffected
	assert.Equal(t, "dev", desc2.Name)
	assert.Equal(t, []string{"t1", "t2"}, desc2.Tags)
}

func TestDescriptorPointerFieldsIndependence(t *testing.T) {
	t.Parallel()

	min := 10.0
	max := 100.0

	d := Descriptor{
		Name: "sensor",
		Min:  &min,
		Max:  &max,
	}

	// Change original values
	min = 20.0
	max = 200.0

	// Descriptor should still point to changed values
	// (showing pointers are shared, not copied)
	assert.Equal(t, 20.0, *d.Min)
	assert.Equal(t, 200.0, *d.Max)
}
