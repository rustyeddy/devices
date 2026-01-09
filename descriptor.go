package devices

// AccessMode indicates the mode of the device, possible modes are
// read-only, write-only and read-write
type AccessMode string

const (
	ReadOnly  AccessMode = "ro"
	WriteOnly AccessMode = "wo"
	ReadWrite AccessMode = "rw"
)

// Described defines the Descriptor interface
type Described interface {
	Descriptor() Descriptor
}

// Descriptor describe the various aspects of the device it
// represents.
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
