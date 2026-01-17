package sensors

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
)

// GPSFix represents a GPS position fix aggregated from NMEA sentences
// (typically GPGGA, GPRMC, and GPVTG). Fields are populated as data becomes
// available from the sensor.
//
// Position coordinates:
//   - Lat, Lon: decimal degrees (negative for South/West)
//   - AltMeters: altitude in meters above mean sea level
//
// Quality indicators:
//   - Quality: GPS fix quality (0=invalid, 1=GPS fix, 2=DGPS fix, etc.)
//   - HDOP: horizontal dilution of precision
//   - Satellites: number of satellites in view
//
// Motion data:
//   - SpeedKnots: ground speed in knots
//   - SpeedMPS: ground speed in meters per second
//   - CourseDeg: course over ground in degrees (0-360, true north)
//
// Status fields:
//   - Status: RMC status ("A" = Active/valid, "V" = Void/invalid)
//   - Date: UTC date in DDMMYY format
type GPSFix struct {
	Lat        float64
	Lon        float64
	AltMeters  float64
	HDOP       float64
	Satellites int
	Quality    int

	SpeedKnots float64
	SpeedMPS   float64
	CourseDeg  float64

	Status string
	Date   string
}

type GTU7Config struct {
	Name    string
	Serial  drivers.SerialConfig
	Factory drivers.SerialFactory

	// Buf sizes the out channel. Default 16.
	Buf int

	// Test injection
	Reader io.Reader
}

type GTU7 struct {
	name string
	cfg  GTU7Config
	out  chan GPSFix
	r    io.ReadCloser
}

func NewGTU7(cfg GTU7Config) *GTU7 {
	if cfg.Factory == nil {
		cfg.Factory = drivers.LinuxSerialFactory{}
	}

	if cfg.Buf <= 0 {
		cfg.Buf = 16
	}

	var r io.ReadCloser
	if cfg.Reader != nil {
		// Wrap test reader with io.NopCloser if it doesn't implement io.Closer
		if rc, ok := cfg.Reader.(io.ReadCloser); ok {
			r = rc
		} else {
			r = io.NopCloser(cfg.Reader)
		}
	} else {
		port, err := cfg.Factory.OpenSerial(cfg.Serial)
		if err != nil {
			panic(err)
		}
		r = port
	}

	return &GTU7{
		name: cfg.Name,
		out:  make(chan GPSFix, cfg.Buf),
		r:    r,
	}
}

func (g *GTU7) Out() <-chan GPSFix { return g.out }

func (g *GTU7) Descriptor() devices.Descriptor {
	attrs := make(map[string]string)
	if g.cfg.Serial.Port != "" {
		attrs["port"] = g.cfg.Serial.Port
	}
	if g.cfg.Serial.Baud > 0 {
		attrs["baud"] = strconv.Itoa(g.cfg.Serial.Baud)
	}

	return devices.Descriptor{
		Name:       g.name,
		Kind:       "gps",
		ValueType:  "GPSFix",
		Access:     devices.ReadOnly,
		Tags:       []string{"gps", "navigation", "location"},
		Attributes: attrs,
	}
}

func (g *GTU7) Run(ctx context.Context) error {
	defer close(g.out)
	defer func() {
		_ = g.r.Close()
	}()

	var last GPSFix
	haveFix := false

	// RMC precedence flags
	haveRMCSpeed := false
	haveRMCCourse := false

	sc := bufio.NewScanner(g.r)
	for sc.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}

		if fix, ok := parseGPGGA(line); ok {
			last.Lat = fix.Lat
			last.Lon = fix.Lon
			last.AltMeters = fix.AltMeters
			last.HDOP = fix.HDOP
			last.Satellites = fix.Satellites
			last.Quality = fix.Quality

			haveFix = true
			g.emit(last)
			continue
		}

		if fix, ok := parseGPRMC(line); ok {
			if !math.IsNaN(fix.Lat) {
				last.Lat = fix.Lat
				last.Lon = fix.Lon
				haveFix = true
			}
			if !math.IsNaN(fix.SpeedKnots) {
				last.SpeedKnots = fix.SpeedKnots
				last.SpeedMPS = fix.SpeedMPS
				haveRMCSpeed = true
			} else {
				haveRMCSpeed = false
			}
			if !math.IsNaN(fix.CourseDeg) {
				last.CourseDeg = fix.CourseDeg
				haveRMCCourse = true
			} else {
				haveRMCCourse = false
			}
			if fix.Status != "" {
				last.Status = fix.Status
			}
			if fix.Date != "" {
				last.Date = fix.Date
			}

			if haveFix {
				g.emit(last)
			}
			continue
		}

		if fix, ok := parseGPVTG(line); ok {
			if !math.IsNaN(fix.SpeedKnots) && (!haveRMCSpeed || math.IsNaN(last.SpeedKnots)) {
				last.SpeedKnots = fix.SpeedKnots
				last.SpeedMPS = fix.SpeedMPS
			}
			if !math.IsNaN(fix.CourseDeg) && (!haveRMCCourse || math.IsNaN(last.CourseDeg)) {
				last.CourseDeg = fix.CourseDeg
			}

			if haveFix {
				g.emit(last)
			}
		}
	}

	return sc.Err()
}

func (g *GTU7) emit(f GPSFix) {
	select {
	case g.out <- f:
	default:
	}
}

/* ---------- Parsing helpers ---------- */

func parseGPGGA(line string) (GPSFix, bool) {
	line = stripChecksum(line)
	parts := strings.Split(line, ",")
	if len(parts) < 10 || (parts[0] != "$GPGGA" && parts[0] != "$GNGGA") {
		return GPSFix{}, false
	}

	lat, lon, err := parseLatLon(parts[2], parts[3], parts[4], parts[5])
	if err != nil {
		return GPSFix{}, false
	}

	q, _ := strconv.Atoi(parts[6])
	sats, _ := strconv.Atoi(parts[7])
	hdop, _ := strconv.ParseFloat(parts[8], 64)
	alt, _ := strconv.ParseFloat(parts[9], 64)

	return GPSFix{
		Lat:        lat,
		Lon:        lon,
		AltMeters:  alt,
		HDOP:       hdop,
		Satellites: sats,
		Quality:    q,
	}, true
}

func parseGPRMC(line string) (GPSFix, bool) {
	line = stripChecksum(line)
	parts := strings.Split(line, ",")
	if len(parts) < 10 || (parts[0] != "$GPRMC" && parts[0] != "$GNRMC") {
		return GPSFix{}, false
	}

	fix := GPSFix{
		Lat:        math.NaN(),
		Lon:        math.NaN(),
		SpeedKnots: math.NaN(),
		SpeedMPS:   math.NaN(),
		CourseDeg:  math.NaN(),
		Status:     parts[2],
		Date:       parts[9],
	}

	if parts[3] != "" {
		if lat, lon, err := parseLatLon(parts[3], parts[4], parts[5], parts[6]); err == nil {
			fix.Lat = lat
			fix.Lon = lon
		}
	}

	if parts[7] != "" {
		if v, err := strconv.ParseFloat(parts[7], 64); err == nil {
			fix.SpeedKnots = v
			fix.SpeedMPS = v * 0.514444
		}
	}

	if parts[8] != "" {
		if v, err := strconv.ParseFloat(parts[8], 64); err == nil {
			fix.CourseDeg = v
		}
	}

	return fix, true
}

func parseGPVTG(line string) (GPSFix, bool) {
	line = stripChecksum(line)
	parts := strings.Split(line, ",")
	if len(parts) < 9 || (parts[0] != "$GPVTG" && parts[0] != "$GNVTG") {
		return GPSFix{}, false
	}

	fix := GPSFix{
		SpeedKnots: math.NaN(),
		SpeedMPS:   math.NaN(),
		CourseDeg:  math.NaN(),
	}

	if parts[1] != "" {
		if v, err := strconv.ParseFloat(parts[1], 64); err == nil {
			fix.CourseDeg = v
		}
	}

	if parts[5] != "" {
		if v, err := strconv.ParseFloat(parts[5], 64); err == nil {
			fix.SpeedKnots = v
			fix.SpeedMPS = v * 0.514444
		}
	}

	return fix, true
}

func stripChecksum(s string) string {
	if i := strings.IndexByte(s, '*'); i >= 0 {
		return s[:i]
	}
	return s
}

func parseLatLon(lat, ns, lon, ew string) (float64, float64, error) {
	if lat == "" || lon == "" {
		return 0, 0, errors.New("latitude or longitude is empty")
	}
	la, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse latitude: %w", err)
	}
	lo, err := strconv.ParseFloat(lon, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse longitude: %w", err)
	}

	latDeg := math.Floor(la / 100)
	latMin := la - latDeg*100
	latVal := latDeg + latMin/60

	lonDeg := math.Floor(lo / 100)
	lonMin := lo - lonDeg*100
	lonVal := lonDeg + lonMin/60

	if ns == "S" {
		latVal = -latVal
	}
	if ew == "W" {
		lonVal = -lonVal
	}
	return latVal, lonVal, nil
}
