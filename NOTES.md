# rig-meter — Notes

## Build and run

Requires Go >= 1.24 and a C compiler for CGO (miniaudio):

    make build
    make test
    make vet

The binary is `rig-meter` (`rig-meter.exe` on Windows).

## Container development

    docker compose build
    docker compose run --rm dev
    docker compose run --rm rig

The container uses the shared PipeWire bootstrap from `scripts/`.
