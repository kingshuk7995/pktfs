# pktfs

This is a go implementation of a programme for transfering files between server and client with server supporting multiple connection with file system level lock.

## Build

### Basic build
```bash
go build -o pktfs ./cmd/pktfs
```

### release build (smaller binary)

```bash
go build -ldflags="-s -w" -o pktfs ./cmd/pktfs
```

### Static build (portable)

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o pktfs ./cmd/pktfs
```

### Cross compilation

#### Linux

```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o pktfs-linux ./cmd/pktfs
```

#### Windows

```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o pktfs.exe ./cmd/pktfs
```

#### macOS

```bash
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o pktfs-mac ./cmd/pktfs
```
