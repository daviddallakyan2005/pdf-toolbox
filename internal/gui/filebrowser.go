package gui

import (
    "os"
    "path/filepath"
    "strings"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"
)

// ShowFileSearchDialog displays a custom file browser with search functionality
func (a *App) ShowFileSearchDialog(callback func([]string), extensions []string, startPath string) {
	if startPath == "" {
		home, _ := os.UserHomeDir()
		startPath = home
	}

	var allFiles []string
	var filteredFiles []string
	currentPath := startPath

    // Search entry
    searchEntry := widget.NewEntry()
    searchEntry.SetPlaceHolder("Type to filter (prefix match)...")

    // Current path label and Up button
    pathLabel := widget.NewLabel(currentPath)
    pathLabel.Wrapping = fyne.TextWrapBreak

	// Forward declare loaders so handlers can call them
    var loadFiles func(string)
    var filterFiles func(string)
    upBtn := widget.NewButton("Up", func() {
        parent := filepath.Dir(currentPath)
        if parent != currentPath {
            currentPath = parent
            pathLabel.SetText(currentPath)
            loadFiles(currentPath)
            searchEntry.SetText("")
        }
    })

    // File list (shows folders and files). Clicking a folder navigates.
    selectedPath := ""
    fileList := widget.NewList(
		func() int { return len(filteredFiles) },
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
        func(i widget.ListItemID, o fyne.CanvasObject) {
            name := filepath.Base(filteredFiles[i])
            if isDir(filteredFiles[i]) {
                name += "/"
            }
            o.(*widget.Label).SetText(name)
		},
	)
    fileList.OnSelected = func(id widget.ListItemID) {
        p := filteredFiles[int(id)]
        if isDir(p) {
            currentPath = p
            pathLabel.SetText(currentPath)
            loadFiles(currentPath)
            searchEntry.SetText("")
            selectedPath = ""
        } else {
            selectedPath = p
        }
    }

    // Load files from directory (includes folders and filtered files)
    loadFiles = func(path string) {
		allFiles = []string{}
		entries, err := os.ReadDir(path)
		if err != nil {
			return
		}

        for _, entry := range entries {
            // Skip hidden entries
            if strings.HasPrefix(entry.Name(), ".") {
                continue
            }

            fullPath := filepath.Join(path, entry.Name())
            if entry.IsDir() {
                allFiles = append(allFiles, fullPath)
                continue
            }

            // Check extension filter for files only
            if len(extensions) > 0 {
                ext := strings.ToLower(filepath.Ext(entry.Name()))
                ok := false
                for _, allowed := range extensions {
                    if ext == strings.ToLower(allowed) {
                        ok = true
                        break
                    }
                }
                if !ok {
                    continue
                }
            }
            allFiles = append(allFiles, fullPath)
        }
		filteredFiles = allFiles
		fileList.Refresh()
	}

    // Filter files based on search (prefix match, case-insensitive)
    filterFiles = func(search string) {
        s := strings.ToLower(search)
        if s == "" {
            filteredFiles = allFiles
        } else {
            filteredFiles = filteredFiles[:0]
            for _, file := range allFiles {
                name := strings.ToLower(filepath.Base(file))
                if strings.HasPrefix(name, s) {
                    filteredFiles = append(filteredFiles, file)
                }
            }
        }
        fileList.Refresh()
    }

	searchEntry.OnChanged = func(text string) {
		filterFiles(text)
	}

	// Change directory button
    // Remove external folder dialog; navigation is in-list via folders and Up button

	// Select button
	selectBtn := widget.NewButton("Select File", func() {
        if selectedPath != "" && !isDir(selectedPath) {
            callback([]string{selectedPath})
        }
	})

	// Layout
    top := container.NewVBox(
        widget.NewLabel("Select a file:"),
        container.NewHBox(upBtn, pathLabel),
        widget.NewSeparator(),
        widget.NewLabel("Search:"),
        searchEntry,
    )
    bottom := container.NewVBox(widget.NewSeparator(), selectBtn)
    content := container.NewBorder(top, bottom, nil, nil, container.NewScroll(fileList))

	// Create and show dialog
	d := dialog.NewCustom("Select File", "Cancel", content, a.window)
	d.Resize(fyne.NewSize(600, 500))

	// Load initial files
	loadFiles(currentPath)

	d.Show()
}

// isDir reports whether path is a directory
func isDir(p string) bool {
    info, err := os.Stat(p)
    if err != nil {
        return false
    }
    return info.IsDir()
}

