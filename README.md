# bpmonitor

`bpmonitor` is a native macOS blood-pressure monitoring application written in
Go. It connects to a GreaterGoods Balance Health 0661 blood-pressure monitor over
Bluetooth Low Energy, decodes its proprietary protocol, persists measurements in
SQLite, and presents a terminal dashboard for reviewing and curating readings.

The project began as an investigation into how to reliably extract useful data
from a specific consumer medical device. It evolved into a complete, testable
application with protocol handling, Bluetooth integration, durable storage, and a
terminal-based review workflow.

## Technical Overview

### BLE Protocol

The monitor does not use the standard Bluetooth Blood Pressure Service. The
project implements the device's proprietary Transtek A6 protocol, including:

- Advertisement-based AES-128 key derivation
- AES-128 ECB frame encryption and decryption
- Custom GATT service and characteristic handling
- Login, time synchronization, measurement, acknowledgement, and follow-up
  indication flows
- Device timestamp conversion and optional measurement metadata

The protocol implementation is isolated from Bluetooth I/O, making frame
construction and parsing deterministic and straightforward to test.

### Application Architecture

```text
cmd/scandebug       BLE discovery, connection, and measurement capture
cmd/bpterm          Terminal dashboard entry point
internal/a6protocol Pure protocol frames, encryption, and parsers
internal/a6session  Handshake state machine and reading assembly
internal/ble        macOS Bluetooth adapter integration
internal/storage    SQLite persistence and review workflow
internal/tui        Bubble Tea dashboard and trends view
```

The protocol and session layers communicate through small data structures and
injected write functions rather than importing the Bluetooth implementation.
This keeps the core behavior platform-independent and allows the handshake to be
tested without physical hardware.

### Storage and Review Workflow

Measurements are stored in SQLite using the pure-Go `modernc.org/sqlite`
driver. The storage layer supports:

- Nullable device metadata
- Duplicate detection for device timestamps
- Morning/night session classification
- Pending, rejected, and committed review states
- Selecting exactly three readings for a sitting
- Atomic commit and cleanup of discarded readings
- Historical trend queries for the dashboard

### Terminal UI

The dashboard is built with Bubble Tea v2, Bubbles v2, Lip Gloss v2, and
NTCharts. It provides a split queue and trends view where readings can be
rejected with undo support, sessions can be committed, and historical systolic
and diastolic values can be filtered by morning or night sessions.

## Engineering Highlights

- Reverse-engineered and validated a proprietary BLE protocol against real
  hardware
- Separated pure protocol logic from platform-specific Bluetooth integration
- Used deterministic unit tests for encryption, parsing, session transitions,
  storage behavior, and TUI state updates
- Designed storage operations around explicit state transitions and transactional
  commit behavior
- Handled CoreBluetooth callback constraints by keeping callbacks lightweight and
  processing BLE events outside callback execution
- Built a hardware-independent demo data path for developing and testing the TUI

## Technology

- Go
- CoreBluetooth through `tinygo.org/x/bluetooth`
- SQLite through `modernc.org/sqlite`
- Bubble Tea v2, Bubbles v2, and Lip Gloss v2
- NTCharts
- AES-128 from Go's standard library

## Project Status

The protocol, session handling, BLE integration, persistence layer, and terminal
dashboard are implemented. The current scope is a focused macOS application for
the GreaterGoods Balance Health 0661 monitor rather than a general-purpose BLE
health-device framework.
