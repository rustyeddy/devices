package devices

// AccessMode indicates read/write capabilities for a device.
type AccessMode string

const (
	// ReadOnly marks devices that can only be read.
	ReadOnly AccessMode = "ro"
	// WriteOnly marks devices that can only be written.
	WriteOnly AccessMode = "wo"
	// ReadWrite marks devices that can be read and written.
	ReadWrite AccessMode = "rw"
)

// Described exposes a Descriptor for a device.
type Described interface {
	Descriptor() Descriptor
}

// Descriptor describes static device metadata.
type Descriptor struct {
	Name       string
	Kind       string // "relay", "button", "temperature", "gps"
	ValueType  string // "bool", "float64", "struct"
	Access     AccessMode
	Unit       string // "C", "%", "rpm", etc.
	Min        *float64
	Max        *float64
	Tags       []string
	Attributes map[string]string // gpio=17, i2c=0x76, etc.
}
