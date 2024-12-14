package parser

import (
	"archive/zip"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/net/html"
)

// Define Lipgloss styles
var (
	h1Style   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#cc241d"))
	h2Style   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#98971a"))
	h3Style   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#b16286"))
	pStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#fff"))
	codeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#458588")).
			Background(lipgloss.Color("#1d2021")).
			Padding(0, 1)
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
				style := getHeaderStyle(headerLevel)
				content := getNodeText(node)
				sb.WriteString(style.Render(content) + "\n\n")
			}
		} else if node.Data == "p" {
			content := getNodeText(node)
			sb.WriteString(pStyle.Render(content) + "\n\n")
		} else if node.Data == "code" {
			content := getNodeText(node)
			sb.WriteString(codeStyle.Render(content) + "\n\n")
		}
	}

	// Traverse child nodes
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		extractText(child, sb)
	}
}

func getHeaderStyle(level int) lipgloss.Style {
	switch level {
	case 1:
		return h1Style
	case 2:
		return h2Style
	case 3:
		return h3Style
	default:
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("60"))
	}
}

func getNodeText(node *html.Node) string {
	var sb strings.Builder
	appendNodeText(node, &sb)
	return strings.TrimSpace(sb.String())
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
