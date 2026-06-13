sin-websearch literature loader for hypothesis refresh.

Invokes `sin-websearch mission` periodically to pull state-of-the-art research and inject verified claims as new hypotheses into `program.md`. If the binary is not on PATH, the loader no-ops silently.

Dependencies: standard library (`context`, `encoding/json`, `fmt`, `os/exec`, `strings`, `time`).
