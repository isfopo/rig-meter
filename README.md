# rig-meter

Standalone capture-level meter for the rig platform.

`rig-meter` is a minimal rig application built from `rig-template`. It opens
the default audio capture device and displays a live level meter in the
terminal.

## Submodules

- `core/` -> `https://github.com/isfopo/rig-core.git`
- `components/` -> `https://github.com/isfopo/rig-components.git`

Clone with submodules:

    git clone --recurse-submodules https://github.com/isfopo/rig-meter.git

## Build (host)

Requires Go >= 1.24 and a C compiler for CGO (RtMidi/miniaudio):

    make build       # build ./rig-meter
    make test        # run all tests (root + submodules)
    make vet         # go vet
    make clean       # remove build artifacts

## Run

    ./rig-meter

Press `q` to quit.

## Layout

    cmd/rig/        application entrypoint
    internal/ui/    meter terminal UI
    core/           rig-core submodule
    components/     rig-components submodule
    scripts/        container entrypoint and PipeWire configuration
