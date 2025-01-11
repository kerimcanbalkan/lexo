package epub

import (
	"archive/zip"
	"fmt"
	"io"
	"strings"
)

func findFile(zipReader *zip.ReadCloser, name string) (io.ReadCloser, error) {
	for _, file := range zipReader.File {
		fmt.Println("file.name den gelen", file.Name)
		if strings.EqualFold(file.Name, name) {
			return file.Open()
		}
	}
	return nil, fmt.Errorf("file not found: %s", name)
}

func readFile(reader io.ReadCloser) ([]byte, error) {
	defer reader.Close()
	return io.ReadAll(reader)
}
