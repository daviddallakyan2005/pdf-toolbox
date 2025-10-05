# Download & Install

Prebuilt binaries are provided on GitHub Releases for macOS, Windows, and Linux (amd64/arm64).

1. Go to the Releases page: [All releases](https://github.com/daviddallakyan2005/pdf-toolbox/releases) or the always-updating [Nightly build](https://github.com/daviddallakyan2005/pdf-toolbox/releases/tag/nightly)
2. Download the archive matching your OS and architecture:
   - macOS: `PDFToolbox-mac-amd64.tar.gz` or `PDFToolbox-mac-arm64.tar.gz`
   - Windows: `PDFToolbox-windows-amd64.zip`
   - Linux: `PDFToolbox-linux-amd64.tar.gz` or `PDFToolbox-linux-arm64.tar.gz`
3. Extract the archive and run the `PDFToolbox` (or `PDFToolbox.exe` on Windows) binary.

Note: On macOS you may need to allow the app in System Settings > Privacy & Security the first time you run it.

## Building From Source (optional)

You can still build locally if you prefer:

```bash
go build -ldflags="-s -w" -o PDFToolbox .
```

This repository includes a GitHub Actions workflow that:
 - publishes a Nightly prerelease on every push to `main`
 - publishes a versioned release when you push a tag like `v1.2.3`
