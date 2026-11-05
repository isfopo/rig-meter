# rig-meter — Specification

## Purpose

`rig-meter` is a standalone capture-level meter. It opens the default audio
capture device and displays a live terminal level meter.

## Dependencies

- `rig-core`: audio engine
- `rig-components`: terminal meter rendering

## Architecture

- `cmd/rig`: application entrypoint
- `internal/ui`: Bubble Tea model and meter view
- `core/`: `rig-core` submodule
- `components/`: `rig-components` submodule

## Current behavior

- Initializes the audio engine
- Opens the default capture device
- Updates a peak meter roughly every 200 ms
- Displays dBFS and frame rate
- Quits on `q` or `ctrl+c`

## Roadmap

- Device selection
- Channel count selection
- Multiple meter styles
- Optional monitor path
