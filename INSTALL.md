# Installation

## Requirements

- Docker installed and running
- macOS, Linux, or Windows

## Download

### macOS (Apple Silicon)

```bash
curl -L https://github.com/stackgen-cli/dataclean/releases/latest/download/dataclean-darwin-arm64.zip -o dataclean.zip
unzip dataclean.zip
chmod +x dataclean
sudo mv dataclean /usr/local/bin/
```

### macOS (Intel)

```bash
curl -L https://github.com/stackgen-cli/dataclean/releases/latest/download/dataclean-darwin-amd64.zip -o dataclean.zip
unzip dataclean.zip
chmod +x dataclean
sudo mv dataclean /usr/local/bin/
```

### Linux (x64)

```bash
curl -L https://github.com/stackgen-cli/dataclean/releases/latest/download/dataclean-linux-amd64.zip -o dataclean.zip
unzip dataclean.zip
chmod +x dataclean
sudo mv dataclean /usr/local/bin/
```

### Windows (x64)

1. Download `dataclean-windows-amd64.zip` from releases
2. Extract `dataclean.exe`
3. Add to PATH or move to a directory in PATH

## Verify Installation

```bash
dataclean --version
```

## From Source

```bash
git clone https://github.com/stackgen-cli/dataclean
cd dataclean
make build
sudo mv dataclean /usr/local/bin/
```

## Uninstall

```bash
sudo rm /usr/local/bin/dataclean
```
