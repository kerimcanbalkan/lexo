package epub

import (
	"encoding/xml"
	"fmt"
	"io"
	"path"
)

type MetaData struct {
	Title       string
	Author      string
	Description string
}

func parseContainer(reader io.Reader) (string, error) {
	type Rootfile struct {
		FullPath string `xml:"full-path,attr"`
	}

	type Container struct {
		Rootfiles struct {
			Rootfile Rootfile `xml:"rootfile"`
		} `xml:"rootfiles"`
	}

	var container Container
	if err := xml.NewDecoder(reader).Decode(&container); err != nil {
		return "", fmt.Errorf("failed to parse container.xml: %v", err)
	}

	return container.Rootfiles.Rootfile.FullPath, nil
}

func parseOPF(reader io.Reader) (MetaData, []string, error) {
	type ManifestItem struct {
		ID        string `xml:"id,attr"`
		Href      string `xml:"href,attr"`
		MediaType string `xml:"media-type,attr"`
	}

	type Manifest struct {
		Items []ManifestItem `xml:"item"`
	}

	type SpineItemRef struct {
		IDRef string `xml:"idref,attr"`
	}

	type Spine struct {
		ItemRefs []SpineItemRef `xml:"itemref"`
	}

	type Package struct {
		Metadata struct {
			Title       string `xml:"title"`
			Creator     string `xml:"creator"`
			Description string `xml:"description"`
		} `xml:"metadata"`
		Manifest Manifest `xml:"manifest"`
		Spine    Spine    `xml:"spine"`
	}

	var pkg Package
	if err := xml.NewDecoder(reader).Decode(&pkg); err != nil {
		return MetaData{}, nil, fmt.Errorf("failed to parse OPF: %v", err)
	}

	metaData := MetaData{
		Title:       pkg.Metadata.Title,
		Author:      pkg.Metadata.Creator,
		Description: pkg.Metadata.Description,
	}

	// Map manifest hrefs by ID
	hrefMap := make(map[string]string)
	for _, item := range pkg.Manifest.Items {
		hrefMap[item.ID] = path.Clean(item.Href)
	}

	// Resolve spine order
	spineOrder := make([]string, 0)
	for _, itemRef := range pkg.Spine.ItemRefs {
		if href, ok := hrefMap[itemRef.IDRef]; ok {
			spineOrder = append(spineOrder, href)
		}
	}
	return metaData, spineOrder, nil
}
