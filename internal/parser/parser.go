package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
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
	pStyle    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "15"})
	codeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "15", Dark: "0"}).
			Background(lipgloss.AdaptiveColor{Light: "0", Dark: "15"})
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
	renderText(body, &sb)
	output := strings.TrimSpace(sb.String())

	return output, nil
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

func renderText(node *html.Node, sb *strings.Builder) {
	switch node.Type {
	case html.TextNode:
		text := strings.TrimSpace(node.Data)
		if text != "" {
			sb.WriteString(text + " ")
		}
	case html.ElementNode:
		switch node.Data {
		case "h1":
			content := getNodeText(node)
			sb.WriteString(h1Style.Render(content) + "\n\n")
		case "h2":
			content := getNodeText(node)
			sb.WriteString(h2Style.Render(content) + "\n")
		case "h3":
			content := getNodeText(node)
			sb.WriteString(h3Style.Render(content) + "\n")
		case "h4":
			content := getNodeText(node)
			sb.WriteString(h4Style.Render(content) + "\n")
		case "h5":
			content := getNodeText(node)
			sb.WriteString(h5Style.Render(content) + "\n")
		case "h6":
			content := getNodeText(node)
			sb.WriteString(h6Style.Render(content) + "\n")
		case "p":
			var inner strings.Builder
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				renderText(child, &inner)
			}
			sb.WriteString(pStyle.Render(inner.String()) + "\n\n")
		case "pre":
			content := getNodeText(node)
			sb.WriteString(lipgloss.NewStyle().Italic(true).Render(content) + "\n\n")
		case "ul":
			l := list.New()
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if child.Type == html.ElementNode && child.Data == "li" {
					content := getNodeText(child)
					l.Item(content)
					// Check for nested lists inside the same li
					for grandchild := child.FirstChild; grandchild != nil; grandchild = grandchild.NextSibling {
						if grandchild.Type == html.ElementNode && grandchild.Data == "ul" {
							renderText(grandchild, sb)
						}
					}
				}
			}
			sb.WriteString(l.String() + "\n\n")
		case "ol":
			l := list.New().Enumerator(list.Roman)
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if child.Type == html.ElementNode && child.Data == "li" {
					content := getNodeText(child)
					l.Item(content)
					// Check for nested lists inside the same li
					for grandchild := child.FirstChild; grandchild != nil; grandchild = grandchild.NextSibling {
						if grandchild.Type == html.ElementNode && grandchild.Data == "ul" {
							renderText(grandchild, sb)
						}
					}
				}
			}
			sb.WriteString(l.String() + "\n\n")
		case "img":
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
		case "strong", "b":
			content := getNodeText(node)
			sb.WriteString(lipgloss.NewStyle().Bold(true).Render(content) + " ")
		case "em", "i":
			content := getNodeText(node)
			sb.WriteString(lipgloss.NewStyle().Italic(true).Render(content) + " ")
		case "code":
			content := getNodeText(node)
			sb.WriteString(codeStyle.Render(content) + "\n\n")
		case "a":
			content := getNodeText(node)
			sb.WriteString(aStyle.Render(content) + " ")
		default:
			// For other inline tags, just recurse
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				renderText(child, sb)
			}
		}
	}
}
