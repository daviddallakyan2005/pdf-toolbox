package gui

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
    "unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
    "github.com/ncruces/zenity"
	"pdf-toolbox/internal/pdf"
	"pdf-toolbox/pkg/models"
)

// App represents the GUI application
type App struct {
	fyneApp    fyne.App
	window     fyne.Window
	pdfService *pdf.Service
}

// NewApp creates a new GUI application
func NewApp() *App {
	return &App{
        fyneApp:    app.NewWithID("com.dallakyan.pdftoolbox"),
		pdfService: pdf.NewService(),
	}
}

// Run starts the application
func (a *App) Run() {
	a.window = a.fyneApp.NewWindow("PDF Toolbox")
	a.window.Resize(fyne.NewSize(800, 600))
	a.window.SetContent(a.makeUI())
	a.window.ShowAndRun()
}

func (a *App) makeUI() fyne.CanvasObject {
	tabs := container.NewAppTabs(
		container.NewTabItem("Split PDF", a.makeSplitTab()),
		container.NewTabItem("Merge PDFs", a.makeMergeTab()),
		container.NewTabItem("Delete Pages", a.makeDeletePagesTab()),
		container.NewTabItem("Images to PDF", a.makeImagesToPDFTab()),
		container.NewTabItem("Info", a.makeInfoTab()),
	)

	return container.NewBorder(
		widget.NewLabel("PDF Toolbox - Split, Merge, and Manage PDFs"),
		nil, nil, nil,
		tabs,
	)
}

func (a *App) makeSplitTab() fyne.CanvasObject {
	var selectedFile string
	fileLabel := widget.NewLabel("No file selected")
	
	pagesEntry := widget.NewEntry()
	pagesEntry.SetPlaceHolder("Pages per file (e.g., 5)")
	pagesEntry.Text = "5"

	outputDirLabel := widget.NewLabel("Output: Same as input file")
	var outputDir string

    selectFileBtn := widget.NewButton("Browse PDF File", func() {
        path, err := a.selectNativeSingle([]zenity.FileFilter{{Name: "PDF", Patterns: []string{"*.pdf"}}})
        if err == nil && path != "" {
            selectedFile = path
            fileLabel.SetText(filepath.Base(selectedFile))
        }
    })
	
	previewBtn := widget.NewButton("Preview PDF", func() {
		if selectedFile == "" {
			dialog.ShowInformation("Preview", "Please select a PDF file first", a.window)
			return
		}
		if err := a.openFile(selectedFile); err != nil {
			dialog.ShowError(err, a.window)
		}
	})

    selectOutputBtn := widget.NewButton("Select Output Directory", func() {
        if dir, err := a.selectNativeFolder(); err == nil && dir != "" {
            outputDir = dir
            outputDirLabel.SetText("Output: " + outputDir)
        }
    })

	splitBtn := widget.NewButton("Split PDF", func() {
		if selectedFile == "" {
			dialog.ShowError(fmt.Errorf("please select a PDF file"), a.window)
			return
		}

		pagesPerFile, err := strconv.Atoi(pagesEntry.Text)
		if err != nil || pagesPerFile < 1 {
			dialog.ShowError(fmt.Errorf("please enter a valid number of pages"), a.window)
			return
		}

		if outputDir == "" {
			outputDir = filepath.Dir(selectedFile)
		}

		config := models.SplitConfig{
			PagesPerFile: pagesPerFile,
			OutputDir:    outputDir,
		}

		go func() {
			err := a.pdfService.Split(config, selectedFile)
			if err != nil {
				dialog.ShowError(err, a.window)
			} else {
				dialog.ShowInformation("Success", "PDF split successfully!", a.window)
			}
		}()
	})

	return container.NewVBox(
		widget.NewLabel("Split a PDF into multiple files"),
		selectFileBtn,
		fileLabel,
		previewBtn,
		widget.NewLabel("Pages per file:"),
		pagesEntry,
		selectOutputBtn,
		outputDirLabel,
		splitBtn,
	)
}

func (a *App) makeMergeTab() fyne.CanvasObject {
    var selectedFiles []string
    var selectedIndex = -1
    selectedMap := map[int]bool{}
    
    var sortable *SortableList
    sortable = NewSortableList(&selectedFiles, &selectedMap, nil)
    sortable.onChange = func(){
        // active row drives preview selection; if none, fall back to last checkbox
        if idx := sortable.ActiveIndex(); idx >= 0 && idx < len(selectedFiles) {
            selectedIndex = idx
        } else {
            selectedIndex = -1
            for i := range selectedFiles { if selectedMap[i] { selectedIndex = i } }
        }
    }
    // ensure list area is visible enough by wrapping in a scroll with min size
    countLabel := widget.NewLabel("0 files selected")
    // track focus index for ↑/↓ based on last selected checkbox
    selectedIndex = -1

    // Add browse with native picker (supports OS search)
    selectFilesBtn := widget.NewButton("Browse & Add PDFs", func() {
        paths, err := a.selectNativeMultiple([]zenity.FileFilter{{Name: "PDF", Patterns: []string{"*.pdf"}}})
        if err == nil && len(paths) > 0 {
            selectedFiles = append(selectedFiles, paths...)
            sortable.rebuild()
            countLabel.SetText(fmt.Sprintf("%d files selected", len(selectedFiles)))
        }
    })
	
	previewBtn := widget.NewButton("Preview Selected", func() {
		if selectedIndex < 0 || selectedIndex >= len(selectedFiles) {
			dialog.ShowInformation("Preview", "Please select a file from the list", a.window)
			return
		}
		if err := a.openFile(selectedFiles[selectedIndex]); err != nil {
			dialog.ShowError(err, a.window)
		}
	})

    clearBtn := widget.NewButton("Clear List", func() {
		selectedFiles = []string{}
        sortable.rebuild()
        countLabel.SetText("0 files selected")
	})

    removeBtn := widget.NewButton("Remove Selected", func() {
        if len(selectedFiles) == 0 { return }
        for i := len(selectedFiles)-1; i >= 0; i-- {
            if selectedMap[i] { selectedFiles = append(selectedFiles[:i], selectedFiles[i+1:]...) }
        }
        selectedMap = map[int]bool{}
        selectedIndex = -1
        sortable.rebuild()
        countLabel.SetText(fmt.Sprintf("%d files selected", len(selectedFiles)))
    })

    moveUpBtn := widget.NewButton("↑", func() {
        i := selectedIndex
        if i <= 0 || i >= len(selectedFiles) {
            return
        }
        selectedFiles[i-1], selectedFiles[i] = selectedFiles[i], selectedFiles[i-1]
        selectedIndex = i - 1
        sortable.rebuild()
    })
    moveDownBtn := widget.NewButton("↓", func() {
        i := selectedIndex
        if i < 0 || i >= len(selectedFiles)-1 {
            return
        }
        selectedFiles[i+1], selectedFiles[i] = selectedFiles[i], selectedFiles[i+1]
        selectedIndex = i + 1
        sortable.rebuild()
    })

    mergeBtn := widget.NewButton("Merge PDFs", func() {
        if len(selectedFiles) < 2 {
            dialog.ShowError(fmt.Errorf("please select at least 2 PDF files"), a.window)
            return
        }
        // If any checkbox is selected, ask whether to merge all or only selected
        hasSelected := false
        var onlySelected []string
        for i, p := range selectedFiles {
            if selectedMap[i] { hasSelected = true; onlySelected = append(onlySelected, p) }
        }
        runMerge := func(inputs []string) {
            if len(inputs) < 2 {
                dialog.ShowError(fmt.Errorf("please select at least 2 files to merge"), a.window)
                return
            }
            suggested := a.suggestMergedName(inputs)
            outputFile, err := a.selectNativeSave(suggested, []zenity.FileFilter{{Name: "PDF", Patterns: []string{"*.pdf"}}})
            if err != nil || outputFile == "" { return }
            config := models.MergeConfig{ InputFiles: inputs, OutputFile: outputFile }
            go func() {
                err := a.pdfService.Merge(config)
                if err != nil { dialog.ShowError(err, a.window) } else { _ = a.openFile(outputFile); dialog.ShowInformation("Success", "PDFs merged successfully!", a.window) }
            }()
        }
        if hasSelected {
            dialog.ShowConfirm("Merge Scope", "Merge only selected files? (No = merge all)", func(only bool) {
                if only { runMerge(onlySelected) } else { runMerge(selectedFiles) }
            }, a.window)
        } else {
            runMerge(selectedFiles)
        }
    })

    listArea := container.NewScroll(sortable.Container())
    listArea.SetMinSize(fyne.NewSize(0, 360))
    return container.NewVBox(
		widget.NewLabel("Merge multiple PDFs into one file"),
        container.NewHBox(selectFilesBtn, previewBtn, removeBtn, moveUpBtn, moveDownBtn),
        countLabel,
		clearBtn,
        listArea,
		mergeBtn,
	)
}

func (a *App) makeDeletePagesTab() fyne.CanvasObject {
	var selectedFile string
	fileLabel := widget.NewLabel("No file selected")
	
	pagesEntry := widget.NewEntry()
	pagesEntry.SetPlaceHolder("Pages to keep (e.g., 1,3,5-7,10)")
    overwrite := false
    overwriteCheck := widget.NewCheck("Overwrite original file", func(v bool) { overwrite = v })

    selectFileBtn := widget.NewButton("Browse PDF File", func() {
        path, err := a.selectNativeSingle([]zenity.FileFilter{{Name: "PDF", Patterns: []string{"*.pdf"}}})
        if err == nil && path != "" {
            selectedFile = path
            fileLabel.SetText(filepath.Base(selectedFile))
            // Get page count
            if count, err := a.pdfService.GetPageCount(selectedFile); err == nil {
                fileLabel.SetText(fmt.Sprintf("%s (%d pages)", filepath.Base(selectedFile), count))
            }
        }
    })
	
	previewBtn := widget.NewButton("Preview PDF", func() {
		if selectedFile == "" {
			dialog.ShowInformation("Preview", "Please select a PDF file first", a.window)
			return
		}
		if err := a.openFile(selectedFile); err != nil {
			dialog.ShowError(err, a.window)
		}
	})

	extractBtn := widget.NewButton("Extract Pages", func() {
		if selectedFile == "" {
			dialog.ShowError(fmt.Errorf("please select a PDF file"), a.window)
			return
		}

		pages, err := parsePageRange(pagesEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid page range: %w", err), a.window)
			return
		}

        var outputFile string
        if overwrite {
            outputFile = selectedFile
        } else {
            suggested := a.suggestDeletedName(selectedFile)
            var err error
            outputFile, err = a.selectNativeSave(suggested, []zenity.FileFilter{{Name: "PDF", Patterns: []string{"*.pdf"}}})
            if err != nil || outputFile == "" {
                return
            }
        }
			config := models.DeletePagesConfig{
				InputFile:   selectedFile,
				OutputFile:  outputFile,
				PagesToKeep: pages,
			}

			go func() {
				err := a.pdfService.DeletePages(config)
				if err != nil {
					dialog.ShowError(err, a.window)
				} else {
                    _ = a.openFile(outputFile)
                    dialog.ShowInformation("Success", "Pages extracted successfully!", a.window)
				}
			}()
	})

    return container.NewVBox(
		widget.NewLabel("Extract specific pages from a PDF"),
		selectFileBtn,
		fileLabel,
		previewBtn,
		widget.NewLabel("Pages to keep:"),
		pagesEntry,
        overwriteCheck,
		extractBtn,
	)
}

func (a *App) makeImagesToPDFTab() fyne.CanvasObject {
	var selectedFiles []string
	var selectedIndex = -1
    selectedMap := map[int]bool{}
	
    var sortable *SortableList
    sortable = NewSortableList(&selectedFiles, &selectedMap, nil)
    sortable.onChange = func(){
        if idx := sortable.ActiveIndex(); idx >= 0 && idx < len(selectedFiles) {
            selectedIndex = idx
        } else {
            selectedIndex = -1
            for i := range selectedFiles { if selectedMap[i] { selectedIndex = i } }
        }
    }
    selectedIndex = -1

    // Add browse with native picker
    selectFilesBtn := widget.NewButton("Browse & Add Images", func() {
        filters := []zenity.FileFilter{{Name: "Images", Patterns: []string{"*.png", "*.jpg", "*.jpeg", "*.gif", "*.bmp"}}}
        paths, err := a.selectNativeMultiple(filters)
        if err == nil && len(paths) > 0 {
            selectedFiles = append(selectedFiles, paths...)
            sortable.rebuild()
        }
    })

    clearBtn := widget.NewButton("Clear List", func() {
		selectedFiles = []string{}
		selectedIndex = -1
        sortable.rebuild()
	})
    removeBtn := widget.NewButton("Remove Selected", func() {
        if len(selectedFiles) == 0 { return }
        for i := len(selectedFiles)-1; i >= 0; i-- { if selectedMap[i] { selectedFiles = append(selectedFiles[:i], selectedFiles[i+1:]...) } }
        selectedMap = map[int]bool{}
        selectedIndex = -1
        sortable.rebuild()
    })
    moveUpBtn := widget.NewButton("↑", func() {
        i := selectedIndex
        if i <= 0 || i >= len(selectedFiles) {
            return
        }
        selectedFiles[i-1], selectedFiles[i] = selectedFiles[i], selectedFiles[i-1]
        selectedIndex = i - 1
        sortable.rebuild()
    })
    moveDownBtn := widget.NewButton("↓", func() {
        i := selectedIndex
        if i < 0 || i >= len(selectedFiles)-1 {
            return
        }
        selectedFiles[i+1], selectedFiles[i] = selectedFiles[i], selectedFiles[i+1]
        selectedIndex = i + 1
        sortable.rebuild()
    })
	
	previewBtn := widget.NewButton("Preview Selected", func() {
		if selectedIndex < 0 || selectedIndex >= len(selectedFiles) {
			dialog.ShowInformation("Preview", "Please select an image from the list", a.window)
			return
		}
		if err := a.openFile(selectedFiles[selectedIndex]); err != nil {
			dialog.ShowError(err, a.window)
		}
	})

    convertBtn := widget.NewButton("Convert to PDF", func() {
        if len(selectedFiles) == 0 {
            dialog.ShowError(fmt.Errorf("please select at least one image file"), a.window)
            return
        }
        // Scope: all or only selected
        hasSelected := false
        var onlySelected []string
        for i, p := range selectedFiles { if selectedMap[i] { hasSelected = true; onlySelected = append(onlySelected, p) } }
        askScope := func(next func(inputs []string)) {
            if hasSelected {
                dialog.ShowConfirm("Convert Scope", "Convert only selected images? (No = convert all)", func(only bool){ if only { next(onlySelected) } else { next(selectedFiles) } }, a.window)
            } else { next(selectedFiles) }
        }
        askScope(func(inputs []string){
            // Mode: single combined or separate
            dialog.ShowConfirm("Conversion Mode", "Combine all into one PDF? (No = separate PDFs)", func(combine bool){
                if combine {
                    outputFile, err := a.selectNativeSave("images.pdf", []zenity.FileFilter{{Name: "PDF", Patterns: []string{"*.pdf"}}})
                    if err != nil || outputFile == "" { return }
                    go func(){ if err := a.pdfService.ImagesToPDF(inputs, outputFile); err != nil { dialog.ShowError(err, a.window) } else { _ = a.openFile(outputFile); dialog.ShowInformation("Success", "Images converted to PDF successfully!", a.window) } }()
                } else {
                    // choose output directory
                    dir, err := a.selectNativeFolder()
                    if err != nil || dir == "" { return }
                    go func(){
                        var lastOut string
                        for _, img := range inputs {
                            name := strings.TrimSuffix(filepath.Base(img), filepath.Ext(img)) + ".pdf"
                            out := filepath.Join(dir, name)
                            if err := a.pdfService.ImagesToPDF([]string{img}, out); err != nil { dialog.ShowError(err, a.window); return }
                            lastOut = out
                        }
                        if lastOut != "" { _ = a.openFile(filepath.Dir(lastOut)) }
                        dialog.ShowInformation("Success", "Images converted to individual PDFs!", a.window)
                    }()
                }
            }, a.window)
        })
    })

    listArea := container.NewScroll(sortable.Container())
    listArea.SetMinSize(fyne.NewSize(0, 360))
    return container.NewVBox(
		widget.NewLabel("Convert multiple images to a single PDF"),
        container.NewHBox(selectFilesBtn, previewBtn, removeBtn, moveUpBtn, moveDownBtn),
        clearBtn,
        listArea,
		convertBtn,
	)
}

func (a *App) makeInfoTab() fyne.CanvasObject {
	var selectedFile string
	fileLabel := widget.NewLabel("No file selected")
	infoLabel := widget.NewLabel("")

    selectFileBtn := widget.NewButton("Browse PDF File", func() {
        path, err := a.selectNativeSingle([]zenity.FileFilter{{Name: "PDF", Patterns: []string{"*.pdf"}}})
        if err == nil && path != "" {
            selectedFile = path
            fileLabel.SetText(filepath.Base(selectedFile))
            if info, err := a.pdfService.GetInfo(selectedFile); err == nil {
                infoText := fmt.Sprintf("File: %v\nPages: %v\nVersion: %v\nSize: %v bytes\nEncrypted: %v",
                    info["file"], info["pages"], info["version"], info["size"], info["encrypted"])
                infoLabel.SetText(infoText)
            } else {
                infoLabel.SetText("Error: " + err.Error())
            }
        }
    })
	
	previewBtn := widget.NewButton("Preview PDF", func() {
		if selectedFile == "" {
			dialog.ShowInformation("Preview", "Please select a PDF file first", a.window)
			return
		}
		if err := a.openFile(selectedFile); err != nil {
			dialog.ShowError(err, a.window)
		}
	})

	return container.NewVBox(
		widget.NewLabel("View PDF information"),
		selectFileBtn,
		fileLabel,
		previewBtn,
		widget.NewSeparator(),
		infoLabel,
	)
}

// parsePageRange parses a page range string like "1,3,5-7,10" into a slice of page numbers
func parsePageRange(rangeStr string) ([]int, error) {
    if strings.TrimSpace(rangeStr) == "" {
        return nil, fmt.Errorf("empty page range")
    }

    // Normalize separators: keep digits and '-' and ','; everything else becomes ','
    normalized := strings.Map(func(r rune) rune {
        if unicode.IsDigit(r) || r == '-' || r == ',' {
            return r
        }
        return ','
    }, rangeStr)

    var pages []int
    for _, raw := range strings.Split(normalized, ",") {
        part := strings.TrimSpace(raw)
        if part == "" {
            continue
        }
        if strings.Contains(part, "-") {
            seg := strings.Split(part, "-")
            if len(seg) != 2 {
                return nil, fmt.Errorf("invalid range: %s", part)
            }
            start, err1 := strconv.Atoi(strings.TrimSpace(seg[0]))
            end, err2 := strconv.Atoi(strings.TrimSpace(seg[1]))
            if err1 != nil || err2 != nil || start < 1 || end < 1 {
                return nil, fmt.Errorf("invalid range: %s", part)
            }
            if start > end {
                return nil, fmt.Errorf("start must be <= end: %s", part)
            }
            for i := start; i <= end; i++ {
                pages = append(pages, i)
            }
            continue
        }
        n, err := strconv.Atoi(part)
        if err != nil || n < 1 {
            return nil, fmt.Errorf("invalid page number: %s", part)
        }
        pages = append(pages, n)
    }

    if len(pages) == 0 {
        return nil, fmt.Errorf("no pages parsed")
    }
    return pages, nil
}

// openFile opens a file in the system's default application
func (a *App) openFile(path string) error {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		return fmt.Errorf("unsupported operating system")
	}
	
	return cmd.Start()
}

// selectNativeSingle opens the OS-native file dialog for a single file.
func (a *App) selectNativeSingle(filters []zenity.FileFilter) (string, error) {
    opts := []zenity.Option{}
    if len(filters) > 0 {
        opts = append(opts, zenity.FileFilters(filters))
    }
    return zenity.SelectFile(opts...)
}

// selectNativeMultiple opens the OS-native file dialog for multiple files.
func (a *App) selectNativeMultiple(filters []zenity.FileFilter) ([]string, error) {
    opts := []zenity.Option{}
    if len(filters) > 0 {
        opts = append(opts, zenity.FileFilters(filters))
    }
    return zenity.SelectFileMultiple(opts...)
}

// selectNativeFolder opens the OS-native folder picker
func (a *App) selectNativeFolder() (string, error) {
    return zenity.SelectFile(zenity.Directory())
}

// selectNativeSave opens the OS-native save dialog
func (a *App) selectNativeSave(defaultName string, filters []zenity.FileFilter) (string, error) {
    opts := []zenity.Option{zenity.ConfirmOverwrite(), zenity.Filename(defaultName)}
    if len(filters) > 0 {
        opts = append(opts, zenity.FileFilters(filters))
    }
    return zenity.SelectFileSave(opts...)
}

// suggestMergedName creates a default filename for merged PDFs based on first and count
func (a *App) suggestMergedName(paths []string) string {
    if len(paths) == 0 {
        return "merged.pdf"
    }
    base := filepath.Base(paths[0])
    name := strings.TrimSuffix(base, filepath.Ext(base))
    if len(paths) == 1 {
        return name + "_merged.pdf"
    }
    return name + fmt.Sprintf("_plus_%d_more_merged.pdf", len(paths)-1)
}

// suggestDeletedName creates a default filename when extracting pages
func (a *App) suggestDeletedName(input string) string {
    base := filepath.Base(input)
    name := strings.TrimSuffix(base, filepath.Ext(base))
    return name + "_pages_extracted.pdf"
}


