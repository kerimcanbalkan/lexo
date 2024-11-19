package epub

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	"github.com/kerimcanbalkan/lexo/internal/parser"
)

// EPUBFile represents an EPUB file
type EPUBFile struct {
	MetaData MetaData
	Contents []string
}

// MetaData represents metadata of an EPUB book
type MetaData struct {
	Title    string `xml:"metadata>title"`
	Language string `xml:"metadata>language"`
	Author   string `xml:"metadata>creator"`
}

// OpenEPUB opens and reads an EPUB file
func OpenEPUB(filePath string) (*EPUBFile, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var epub EPUBFile
	for _, f := range r.File {
		// Check for metadata and content files
		if strings.HasSuffix(f.Name, ".opf") {
			// Metadata file found
			meta, err := ReadMetaData(f)
			if err != nil {
				return nil, err
			}
			epub.MetaData = meta
		} else if strings.HasSuffix(f.Name, ".html") || strings.HasSuffix(f.Name, ".xhtml") {
			// Add content files
			c, err := readContentFile(f)
			if err != nil {
				return nil, err
			}
			epub.Contents = append(epub.Contents, c)
		}
	}
	return &epub, nil
}

// readMetaData reads metadata from the .opf file
func ReadMetaData(f *zip.File) (MetaData, error) {
	rc, err := f.Open()
	if err != nil {
		return MetaData{}, err
	}
	defer rc.Close()

	var meta MetaData
	decoder := xml.NewDecoder(rc)
	err = decoder.Decode(&meta)
	return meta, err
}

func readContentFile(f *zip.File) (string, error) {
	parsedContent, err := parser.ParseHTML(f)
	if err != nil {
		fmt.Println("could not parse the html", err)
		os.Exit(1)
	}
	return parsedContent, nil
}
