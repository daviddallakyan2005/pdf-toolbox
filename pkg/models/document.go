package models

// Document represents a PDF or image file
type Document struct {
	Path      string
	Name      string
	PageCount int
	Type      DocumentType
}

// DocumentType represents the type of document
type DocumentType int

const (
	TypePDF DocumentType = iota
	TypeImage
)

// Operation represents a PDF operation type
type Operation int

const (
	OpSplit Operation = iota
	OpMerge
	OpDeletePages
	OpPreview
)

// SplitConfig holds configuration for splitting PDFs
type SplitConfig struct {
	PagesPerFile int
	OutputDir    string
}

// MergeConfig holds configuration for merging PDFs
type MergeConfig struct {
	InputFiles []string
	OutputFile string
}

// DeletePagesConfig holds configuration for deleting pages
type DeletePagesConfig struct {
	InputFile   string
	OutputFile  string
	PagesToKeep []int
}

