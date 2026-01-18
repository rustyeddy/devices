
---

## Why this is a *good canonical example*

- **Pure application logic**  
  No OttO runtime coupling — just devices + drivers

- **Matches your architectural goal**  
  Apps embed devices, not the other way around

- **Codex-friendly**  
  This file is an ideal pattern for:
  - future examples
  - tests
  - irrigation controller apps
  - docs generation

---

## Optional next refinements (when you’re ready)

1. Add a **mock GPIO driver** version:
   ```bash
   go run ./examples/led --mock
