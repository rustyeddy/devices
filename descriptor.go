package devices

type AccessMode string

const (
	ReadOnly  AccessMode = "ro"
	WriteOnly AccessMode = "wo"
	ReadWrite AccessMode = "rw"
)

type Described interface {
	Descriptor() Descriptor
}

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
