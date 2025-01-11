package epub

import (
	"archive/zip"
	"fmt"

	"github.com/kerimcanbalkan/lexo/internal/parser"
)

type EPUBFile struct {
	MetaData MetaData
	Contents []string
}

type EPUBReader struct {
	filePath  string
	zipReader *zip.ReadCloser
}

func NewEPUBReader(filePath string) (*EPUBReader, error) {
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB file: %v", err)
	}
	return &EPUBReader{filePath: filePath, zipReader: zipReader}, nil
}

func (r *EPUBReader) Parse() (*EPUBFile, error) {
	defer r.zipReader.Close()

	// Locate container.xml
	containerPath := "META-INF/container.xml"
	containerFile, err := findFile(r.zipReader, containerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to locate container.xml: %v", err)
	}

	// Parse container.xml to find the OPF path
	opfPath, err := parseContainer(containerFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse container.xml: %v", err)
	}

	// Locate and parse OPF file
	opfFile, err := findFile(r.zipReader, opfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to locate OPF file: %v", err)
	}

	metaData, spine, err := parseOPF(opfFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OPF file: %v", err)
	}

	// Read the XHTML contents in the spine order
	contents := make([]string, 0)
	for _, href := range spine {
		contentFile, err := findFile(r.zipReader, "OEBPS/"+href)
		if err != nil {
			continue // Skip missing files
		}
		contentData, err := parser.ParseHTML(contentFile)
		if err != nil {
			continue // Skip unreadable files
		}
		contents = append(contents, string(contentData))
	}

	return &EPUBFile{MetaData: metaData, Contents: contents}, nil
}
