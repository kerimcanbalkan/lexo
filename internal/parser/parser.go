package parser

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/net/html"
)

// Define Lipgloss styles
var (
	h1Style = lipgloss.NewStyle().
		Bold(true).
		SetString("*").
		Foreground(lipgloss.AdaptiveColor{Light: "2", Dark: "11"})
	h2Style = lipgloss.NewStyle().
		Bold(true).
		SetString("**").
		Foreground(lipgloss.AdaptiveColor{Light: "4", Dark: "13"})
	h3Style = lipgloss.NewStyle().
		SetString("***").
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "1", Dark: "1"})
	h4Style = lipgloss.NewStyle().
		SetString("****").
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "13", Dark: "12"})
	h5Style = lipgloss.NewStyle().
		SetString("*****").
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "9", Dark: "6"})
	h6Style = lipgloss.NewStyle().
		SetString("******").
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "14", Dark: "11"})
	aStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "4", Dark: "4"}).
		Underline(true)
	pStyle  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "15"})
	liStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "4", Dark: "4"}).SetString("•").Padding(0, 1)
	codeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "15", Dark: "0"}).
			Background(lipgloss.AdaptiveColor{Light: "0", Dark: "15"}).
			Padding(0, 1)
)

func ParseHTML(f io.ReadCloser) (string, error) {
	defer f.Close()

	doc, err := html.Parse(f)
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
				sb.WriteString(style.Render(content) + "\n")
			}
		} else if node.Data == "p" {
			var inner strings.Builder
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				renderInline(child, &inner)
			}
			sb.WriteString(pStyle.Render(inner.String()) + "\n\n")
		} else if node.Data == "code" {
			content := getNodeText(node)
			sb.WriteString(codeStyle.Render(content) + "\n")
		} else if node.Data == "pre" {
			content := getNodeText(node)
			sb.WriteString(codeStyle.Render(content) + "\n")
		} else if node.Data == "a" {
			content := getNodeText(node)
			sb.WriteString(aStyle.Render(content) + "\n")
		} else if node.Data == "ul" {
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if child.Type == html.ElementNode && child.Data == "li" {
					content := getNodeText(child)
					sb.WriteString(liStyle.Render(content) + "\n")
					// Check for nested lists inside the same li
					for grandchild := child.FirstChild; grandchild != nil; grandchild = grandchild.NextSibling {
						if grandchild.Type == html.ElementNode && grandchild.Data == "ul" {
							extractText(grandchild, sb)
						}
					}
				}
			}
			sb.WriteString("\n")
		} else if node.Data == "img" {
			var altText string

			// Extract the alt attribute
			for _, attr := range node.Attr {
				if attr.Key == "alt" {
					altText = attr.Val
					break
				}
			}

			if altText != "" {
				sb.WriteString(pStyle.Render(fmt.Sprintf("[IMG: %s]", altText)) + "\n\n")
			} else {
				sb.WriteString(pStyle.Render("[IMG]") + "\n\n")
			}
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
	case 4:
		return h4Style
	case 5:
		return h5Style
	case 6:
		return h6Style
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

func renderInline(node *html.Node, sb *strings.Builder) {
	switch node.Type {
	case html.TextNode:
		text := strings.TrimSpace(node.Data)
		if text != "" {
			sb.WriteString(text + " ")
		}
	case html.ElementNode:
		switch node.Data {
		case "strong", "b":
			content := getNodeText(node)
			sb.WriteString(lipgloss.NewStyle().Bold(true).Render(content))
		case "em", "i":
			content := getNodeText(node)
			sb.WriteString(lipgloss.NewStyle().Italic(true).Render(content))
		case "code":
			content := getNodeText(node)
			sb.WriteString(codeStyle.Render(content))
		case "a":
			content := getNodeText(node)
			sb.WriteString(aStyle.Render(content))
		default:
			// For other inline tags, just recurse
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				renderInline(child, sb)
			}
		}
	}
}
