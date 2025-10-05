package pdf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"pdf-toolbox/internal/utils"
	"pdf-toolbox/pkg/models"
)

// Service handles PDF operations
type Service struct{}

// NewService creates a new PDF service
func NewService() *Service {
	return &Service{}
}

// GetPageCount returns the number of pages in a PDF
func (s *Service) GetPageCount(filePath string) (int, error) {
	ctx, err := api.ReadContextFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read PDF: %w", err)
	}
	return ctx.PageCount, nil
}

// Split splits a PDF into multiple files with specified pages per file
func (s *Service) Split(config models.SplitConfig, inputFile string) error {
	if !utils.IsPDF(inputFile) {
		return fmt.Errorf("input file must be a PDF")
	}

	if err := utils.EnsureDir(config.OutputDir); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get total page count
	pageCount, err := s.GetPageCount(inputFile)
	if err != nil {
		return err
	}

	// Split the PDF
	baseName := filepath.Base(inputFile)
	baseName = baseName[:len(baseName)-len(filepath.Ext(baseName))]

	for start := 1; start <= pageCount; start += config.PagesPerFile {
		end := start + config.PagesPerFile - 1
		if end > pageCount {
			end = pageCount
		}

		outputFile := filepath.Join(config.OutputDir, fmt.Sprintf("%s_pages_%d-%d.pdf", baseName, start, end))
		
		span := fmt.Sprintf("%d-%d", start, end)
		if err := api.TrimFile(inputFile, outputFile, []string{span}, nil); err != nil {
			return fmt.Errorf("failed to split pages %d-%d: %w", start, end, err)
		}
	}

	return nil
}

// Merge combines multiple PDF files into one
func (s *Service) Merge(config models.MergeConfig) error {
	if len(config.InputFiles) == 0 {
		return fmt.Errorf("no input files provided")
	}

	for _, file := range config.InputFiles {
		if !utils.IsPDF(file) {
			return fmt.Errorf("all input files must be PDFs: %s", file)
		}
	}

	if err := utils.EnsureDir(filepath.Dir(config.OutputFile)); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return api.MergeCreateFile(config.InputFiles, config.OutputFile, false, nil)
}

// DeletePages removes specified pages from a PDF
func (s *Service) DeletePages(config models.DeletePagesConfig) error {
	if !utils.IsPDF(config.InputFile) {
		return fmt.Errorf("input file must be a PDF")
	}

	if len(config.PagesToKeep) == 0 {
		return fmt.Errorf("must specify pages to keep")
	}

	if err := utils.EnsureDir(filepath.Dir(config.OutputFile)); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

    // Build page selectors as individual tokens to avoid downstream parsing issues
    selectors := make([]string, 0, len(config.PagesToKeep))
    for _, page := range config.PagesToKeep {
        selectors = append(selectors, fmt.Sprintf("%d", page))
    }

    return api.TrimFile(config.InputFile, config.OutputFile, selectors, nil)
}

// ExtractPages extracts specific pages to a new PDF
func (s *Service) ExtractPages(inputFile, outputFile string, pages []int) error {
	if !utils.IsPDF(inputFile) {
		return fmt.Errorf("input file must be a PDF")
	}

	if len(pages) == 0 {
		return fmt.Errorf("no pages specified")
	}

	if err := utils.EnsureDir(filepath.Dir(outputFile)); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

    // Build selectors as individual tokens
    selectors := make([]string, 0, len(pages))
    for _, page := range pages {
        selectors = append(selectors, fmt.Sprintf("%d", page))
    }

    return api.TrimFile(inputFile, outputFile, selectors, nil)
}

// ImagesToPDF converts multiple images to a single PDF
func (s *Service) ImagesToPDF(imageFiles []string, outputFile string) error {
	if len(imageFiles) == 0 {
		return fmt.Errorf("no image files provided")
	}

	if err := utils.EnsureDir(filepath.Dir(outputFile)); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create import configuration
	imp := pdfcpu.DefaultImportConfig()

	return api.ImportImagesFile(imageFiles, outputFile, imp, nil)
}

// GetInfo returns information about a PDF file
func (s *Service) GetInfo(filePath string) (map[string]interface{}, error) {
	if !utils.IsPDF(filePath) {
		return nil, fmt.Errorf("file must be a PDF")
	}

	ctx, err := api.ReadContextFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %w", err)
	}

	info := map[string]interface{}{
		"pages":     ctx.PageCount,
		"version":   ctx.HeaderVersion,
		"file":      filepath.Base(filePath),
		"size":      getFileSize(filePath),
		"encrypted": ctx.Encrypt != nil,
	}

	return info, nil
}

func getFileSize(filePath string) int64 {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return info.Size()
}

