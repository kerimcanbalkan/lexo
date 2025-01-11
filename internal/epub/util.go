package epub

import (
	"archive/zip"
	"fmt"
	"io"
	"strings"
)

func findFile(zipReader *zip.ReadCloser, name string) (io.ReadCloser, error) {
	for _, file := range zipReader.File {
		if strings.EqualFold(file.Name, name) || strings.HasSuffix(file.Name, name) {
			return file.Open()
		}
	}
	return nil, fmt.Errorf("file not found: %s", name)
}

func readFile(reader io.ReadCloser) ([]byte, error) {
	defer reader.Close()
	return io.ReadAll(reader)
}
