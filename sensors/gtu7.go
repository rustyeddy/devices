package sensors

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
)

// GPSFix is a normalized position fix derived from NMEA sentences.
//
// Lat/Lon are decimal degrees (W/S are negative).
// AltitudeM is meters above mean sea level (per GGA).
type GPSFix struct {
	Lat       float64
	Lon       float64
	AltitudeM float64
	Quality   int
	Satellites int
	HDOP      float64
	UTCTime   string // HHMMSS.SS as emitted by the receiver
}

// GTU7Config configures a Vegetronix GT-U7 GPS (NMEA) reader.
//
// Provide either Reader (for tests) or Factory+Serial (for real hardware).
type GTU7Config struct {
	Name    string
	Reader  io.Reader
	Factory drivers.SerialFactory
	Serial  drivers.SerialConfig
	OutBuf  int
}

// GTU7 reads NMEA sentences from a serial port and emits GPS fixes.
type GTU7 struct {
	devices.Base
	cfg   GTU7Config
	out   chan GPSFix
}

func NewGTU7(cfg GTU7Config) *GTU7 {
	if cfg.Name == "" {
		cfg.Name = "gtu7"
	}
	if cfg.OutBuf <= 0 {
		cfg.OutBuf = 8
	}
	return &GTU7{
		Base: devices.NewBase(cfg.Name, 16),
		cfg:  cfg,
		out:  make(chan GPSFix, cfg.OutBuf),
	}
}

func (g *GTU7) Out() <-chan GPSFix { return g.out }

func (g *GTU7) Descriptor() devices.Descriptor {
	return devices.Descriptor{
		Name:      g.Name(),
		Kind:      "gps",
		ValueType: "gps_fix",
		Access:    devices.ReadOnly,
		Unit:      "",
		Tags:      []string{"location", "nmea"},
		Attributes: map[string]string{
			"serial": g.cfg.Serial.Port,
			"baud":   fmt.Sprintf("%d", g.cfg.Serial.Baud),
		},
	}
}

func (g *GTU7) Run(ctx context.Context) error {
	defer close(g.out)
	defer g.CloseEvents()

	var (
		r   io.Reader
		sp  drivers.SerialPort
		err error
	)

	if g.cfg.Reader != nil {
		r = g.cfg.Reader
		g.Emit(devices.EventInfo, "using provided reader", nil, nil)
	} else {
		factory := g.cfg.Factory
		if factory == nil {
			factory = drivers.LinuxSerialFactory{}
		}
		sp, err = factory.OpenSerial(g.cfg.Serial)
		if err != nil {
			g.EmitBlocking(devices.EventError, "failed to open serial", err, map[string]string{"port": g.cfg.Serial.Port})
			return err
		}
		defer sp.Close()
		r = sp
		g.Emit(devices.EventInfo, "serial opened", nil, map[string]string{"port": sp.String()})
	}

	scanner := bufio.NewScanner(r)
	// NMEA lines are small; default 64K buffer is plenty, but allow a bit more.
	buf := make([]byte, 0, 4096)
	scanner.Buffer(buf, 64*1024)

	for {
		select {
		case <-ctx.Done():
			g.Emit(devices.EventInfo, "stopping", nil, nil)
			return nil
		default:
		}

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				g.Emit(devices.EventError, "scanner error", err, nil)
				return err
			}
			// EOF.
			return nil
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "$") {
			continue
		}

		if fix, ok := parseGPGGA(line); ok {
			select {
			case g.out <- fix:
			default:
				// drop if slow consumer
			}
		}
	}
}

// parseGPGGA extracts a fix from a $GPGGA sentence.
// Returns ok=false for non-GPGGA lines or parse failures.
func parseGPGGA(line string) (GPSFix, bool) {
	// Strip checksum, if present.
	// We keep parsing even if checksum is present/absent; validation is out of scope for now.
	if i := strings.IndexByte(line, '*'); i >= 0 {
		line = line[:i]
	}

	parts := strings.Split(line, ",")
	if len(parts) < 10 {
		return GPSFix{}, false
	}
	if parts[0] != "$GPGGA" && parts[0] != "$GNGGA" {
		return GPSFix{}, false
	}

	utc := parts[1]
	latRaw, latHem := parts[2], parts[3]
	lonRaw, lonHem := parts[4], parts[5]
	quality, _ := strconv.Atoi(parts[6])
	sats, _ := strconv.Atoi(parts[7])
	hdop, _ := strconv.ParseFloat(parts[8], 64)
	alt, _ := strconv.ParseFloat(parts[9], 64)

	lat, ok := nmeaDeg(latRaw)
	if !ok {
		return GPSFix{}, false
	}
	if strings.EqualFold(latHem, "S") {
		lat = -lat
	}

	lon, ok := nmeaDeg(lonRaw)
	if !ok {
		return GPSFix{}, false
	}
	if strings.EqualFold(lonHem, "W") {
		lon = -lon
	}

	// quality 0 means invalid fix.
	if quality == 0 {
		return GPSFix{}, false
	}

	return GPSFix{
		Lat:        lat,
		Lon:        lon,
		AltitudeM:  alt,
		Quality:    quality,
		Satellites: sats,
		HDOP:       hdop,
		UTCTime:    utc,
	}, true
}

// nmeaDeg converts ddmm.mmmm (lat) or dddmm.mmmm (lon) into decimal degrees.
func nmeaDeg(v string) (float64, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, false
	}
	deg := math.Floor(f / 100.0)
	min := f - (deg * 100.0)
	return deg + (min / 60.0), true
}
