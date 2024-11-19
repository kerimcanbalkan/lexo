package parser

import (
	"archive/zip"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func ParseHTML(f *zip.File) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	doc, err := html.Parse(rc)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %v", err)
	}

	// Find the body tag
	body := findBody(doc)
	if body == nil {
		return "", fmt.Errorf("no <body> tag found")
	}

	var sb strings.Builder
	extractText(body, &sb)
	return sb.String(), nil
}

func findBody(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.Data == "body" {
		return node
	}

	// Traverse child nodes
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if result := findBody(child); result != nil {
			return result
		}
	}
	return nil
}

func extractText(node *html.Node, sb *strings.Builder) {
	if node.Type == html.ElementNode {
		// Handle header tags dynamically (h1 through h6)
		if strings.HasPrefix(node.Data, "h") {
			headerLevel, err := strconv.Atoi(node.Data[1:])
			if err == nil && headerLevel >= 1 && headerLevel <= 6 {
				sb.WriteString(strings.Repeat("#", headerLevel))
				appendNodeText(node, sb) // Append the header content
				sb.WriteString(strings.Repeat("#", headerLevel) + "\n")
			}
		} else if node.Data == "p" {
			appendNodeText(node, sb) // Append paragraph content
			sb.WriteString("\n")
		}
	}

	// Traverse child nodes
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		extractText(child, sb)
	}
}

func appendNodeText(node *html.Node, sb *strings.Builder) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode {
			text := strings.TrimSpace(child.Data)
			if len(text) > 0 {
				sb.WriteString(text + " ")
			}
		}
		// Traverse deeper if the child node contains other elements with text
		appendNodeText(child, sb)
	}
}
