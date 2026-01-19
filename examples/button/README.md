# Button Example (GPIOCDev + Events)

This example demonstrates a GPIO button using the modern `devices` API.

It shows **both** ways to observe a button:

1. **State stream** via `Out()` (`<-chan bool`)
2. **Event stream** via `Events()` (`<-chan devices.Event`)

---

## What this example shows

- Using `drivers.NewGPIOCDevFactory()` (Linux GPIO character device API)
- Creating a button with `button.ButtonConfig`
- Running the button with a `context.Context`
- Reading boolean state transitions from `Out()`
- Reading lifecycle + edge notifications from `Events()`
- Interpreting states as `PRESSED` / `RELEASED`
- Clean shutdown via OS signals

---

## Requirements

- Linux kernel with GPIO character device support (`/dev/gpiochip*`)
- A momentary pushbutton wired to a GPIO line
- A ground reference shared with the GPIO system

> Note: This example defaults to `BiasPullUp`, which is common: the GPIO line is
> normally HIGH (released), and goes LOW when the button is pressed (wired to GND).

---

## Hardware assumptions

Defaults in `main.go`:

- GPIO chip: `gpiochip0`
- GPIO line offset: `23` (change to match your wiring)
- Bias: `PullUp`
- Edge: `Both` (rising + falling)
- Debounce: `30ms`

With `PullUp` bias:

- **Released**: state = `true` (HIGH)
- **Pressed**:  state = `false` (LOW)

If you wire your button differently (pull-down), change:

- `Bias` to `drivers.BiasPullDown`
- (and the pressed/released interpretation flips automatically)

---

## Finding the correct GPIO line

List available GPIO chips:

```bash
gpiodetect
