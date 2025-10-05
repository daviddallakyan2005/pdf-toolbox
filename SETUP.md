# Build Instructions

## macOS

### ARM64 (M1/M2/M3)
```bash
GOARCH=arm64 go build -ldflags="-s -w" -o PDFToolbox-mac-arm64 .
```

### Intel (x86_64)
```bash
GOARCH=amd64 go build -ldflags="-s -w" -o PDFToolbox-mac-amd64 .
```

## Windows

**Build on Windows machine:**
```bash
go build -ldflags="-s -w -H windowsgui" -o PDFToolbox.exe .
```

## Linux

### x86_64
```bash
go build -ldflags="-s -w" -o PDFToolbox-linux-amd64 .
```

### ARM64
```bash
GOARCH=arm64 go build -ldflags="-s -w" -o PDFToolbox-linux-arm64 .
```

---

**Note:** Cross-compilation for GUI apps with OpenGL requires building on the target OS. Build Windows .exe on Windows, macOS binaries on macOS, Linux binaries on Linux.

**Flags:**
- `-s -w` = strip debug info (smaller binary)
- `-H windowsgui` = hide console window on Windows

