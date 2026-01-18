# Relay Example (GPIOCDev)

This example demonstrates how to drive a relay using the **modern `devices` API**
with Linux **GPIO character devices (GPIOCDev)**.

It mirrors the LED example and is intended as the **canonical actuator example**
for relays in this repository.

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
- A relay module wired to a GPIO line
- Proper power wiring for the relay module (many need 5V + GND)

> ⚠️ Relays switch real loads. Be safe. If switching mains voltage, use proper
> enclosures, fusing, strain relief, and isolation. When in doubt, don’t.

---

## Hardware assumptions

- GPIO chip: `gpiochip0`
- GPIO line offset: `18` (change this to match your wiring)
- Relay control signal connected to GPIO line and ground reference shared

Some relay boards are **active-low** (GPIO low turns relay ON). If your relay is
active-low, your device config may include a field like `Invert` or `ActiveLow`.
If present, set it accordingly in `RelayConfig`.

---

## Finding the correct GPIO line

List available GPIO chips:

```bash
gpiodetect
```
