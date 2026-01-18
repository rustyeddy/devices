# LED Example (GPIOCDev)

This example demonstrates how to drive a simple LED using the **modern `devices` API**
with Linux **GPIO character devices (GPIOCDev)**.

It is intended as the **canonical actuator example** for this repository.

---

## What this example shows

- Using the `drivers.GPIOCDevFactory` (Linux GPIO character device API)
- Creating a device using a config struct
- Running a device with a `context.Context`
- Sending commands over a **non-blocking channel**
- Clean shutdown via OS signals

---

## Requirements

- Linux kernel with GPIO character device support (`/dev/gpiochip*`)
- `libgpiod` compatible kernel (modern Linux)
- LED wired with a **current-limiting resistor**

---

## Hardware assumptions

- GPIO chip: `gpiochip0`
- GPIO line offset: `17`
- LED connected between GPIO line and ground

> ⚠️ Always use a resistor (typically 220Ω–1kΩ) to avoid damaging the GPIO pin.

---

## Finding the correct GPIO line

List available GPIO chips:

```bash
gpiodetect

Inspect lines on a chip:

gpioinfo gpiochip0


Look for a free line number (offset) you can safely use.
