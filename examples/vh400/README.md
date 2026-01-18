# VH400 Example (ADS1115 via periph.io)

This example demonstrates reading a Vegetronix **VH400** soil moisture sensor
(volumetric water content, **VWC %**) using the modern `devices` API.

The VH400 outputs an **analog voltage**, so this device expects an ADC. This example
uses an **ADS1115** connected over I2C.

---

## What this example shows

- Using an `ADCFactory` (`drivers.PeriphADCFactory{}`) to open an ADS1115
- Creating the VH400 sensor via `vh400.VH400Config`
- Running the device with a `context.Context`
- Reading samples from `Out()` and printing VWC (%)
- Clean shutdown via OS signals

---

## Requirements

- VH400 sensor wired to an ADS1115 input (A0–A3)
- ADS1115 wired to I2C (SDA/SCL + 3.3V + GND)
- Linux device with I2C enabled

### Important build note (repo-specific)

In this repository, the ADS1115 periph-backed implementation is built only on:

- `linux` **and** (`arm` or `arm64`)

So this example is intended to run on a Raspberry Pi (or similar).
On other platforms, the ADS1115 implementation is stubbed.

---

## Hardware assumptions

Defaults in `main.go`:

- I2C bus: `"1"` (common on Raspberry Pi)
- ADS1115 address: `0x48`
- ADS1115 channel: `0` (single-ended A0)
- Sample interval: 2 seconds

If your wiring differs, update:

- `Bus`, `Addr`, and `Channel`

---

## Enable I2C (Raspberry Pi notes)

You must have I2C enabled and the device visible.

Common checks:

```bash
ls /dev/i2c-*
i2cdetect -y 1
```

You should see 48 if the ADS1115 is detected at 0x48.
