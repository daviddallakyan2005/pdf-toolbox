# PDF Toolbox

> Lightweight cross-platform PDF utility: split, merge, extract pages, and convert images to PDF.

## Features

- **Split PDFs** – divide by page count (e.g., 5 pages per file)
- **Merge PDFs** – combine multiple PDFs into one
- **Extract Pages** – keep specific pages (e.g., `1,3,5-7,10`)
- **Images to PDF** – convert PNG/JPG/JPEG/GIF/BMP to PDF
- **PDF Info** – view page count, version, size, encryption status
- **File Search** – custom file browser with real-time search filtering (type to filter files by name)
- **Preview** – open PDFs and images in system viewer
- **Folder Navigation** – easily switch between directories to find your files

## Quick Start

### macOS
```bash
# Install dependencies (one-time)
xcode-select --install  # if needed

# Clone and build
git clone https://github.com/daviddallakyan2005/pdf-toolbox.git
cd pdf-toolbox
go mod download
go build -o PDFToolbox .

# Run
./PDFToolbox
```

### Windows
```bash
# Clone and build
git clone https://github.com/daviddallakyan2005/pdf-toolbox.git
cd pdf-toolbox
go mod download
go build -o PDFToolbox.exe .

# Run
PDFToolbox.exe
```

**Cross-compile Windows .exe on Mac:**
```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o PDFToolbox.exe .
```

### Linux
```bash
# Install dependencies (one-time)
sudo apt install gcc libgl1-mesa-dev xorg-dev  # Debian/Ubuntu

# Clone and build
git clone https://github.com/daviddallakyan2005/pdf-toolbox.git
cd pdf-toolbox
go mod download
go build -o PDFToolbox .

# Run
./PDFToolbox
```

## Requirements

- **Go 1.21+**
- **macOS:** Xcode command line tools
- **Windows:** None (static binary)
- **Linux:** gcc, libgl1-mesa-dev, xorg-dev

## Usage

1. **Split PDF:** Select file → enter pages per output (default: 5) → choose output folder → Split
2. **Merge PDFs:** Select multiple files (click repeatedly) → Merge → save output
3. **Extract Pages:** Select file → enter pages to keep (e.g., `1,3,5-7`) → Extract → save
4. **Images to PDF:** Select images → Convert → save
5. **Info:** Select PDF → view details

## Project Structure

```
pdf-toolbox/
├── main.go              # Entry point
├── internal/
│   ├── gui/app.go       # Fyne GUI
│   ├── pdf/operations.go # PDF operations
│   └── utils/fileutils.go
└── pkg/models/document.go
```

## Tech Stack

- [Go](https://golang.org/) – language
- [Fyne](https://fyne.io/) – GUI framework
- [pdfcpu](https://github.com/pdfcpu/pdfcpu) – PDF engine

## Background

Created to solve AUA students' printing restrictions (5 pages max per job). Now a full PDF management tool.

---

**License:** MIT
