package main

// An example program demonstrating the pager component from the Bubbles
// component library.

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kerimcanbalkan/lexo/internal/epub"
)

// You generally won't need this unless you're processing stuff with
// complicated ANSI escape sequences. Turn it on if you notice flickering.
//
// Also keep in mind that high performance rendering only works for programs
// that use the full size of the terminal. We're enabling that below with
// tea.EnterAltScreen().
const useHighPerformanceRenderer = false

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()
	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

type model struct {
	content  string
	title    string
	ready    bool
	viewport viewport.Model
}

func (m *model) loadYOffset() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// can't get config dir → default to zero and log for visibility
		m.viewport.YOffset = 0
		return
	}

	appDir := filepath.Join(configDir, "lexo")
	name := sanitize(m.title) + ".pos"
	posPath := filepath.Join(appDir, name)

	data, err := os.ReadFile(posPath)
	if err != nil {
		// file missing or unreadable → default to zero (likely first run)
		m.viewport.YOffset = 0
		return
	}

	s := strings.TrimSpace(string(data))
	posNum, err := strconv.Atoi(s)
	if err != nil {
		// corrupted data → fallback to zero
		m.viewport.YOffset = 0
		return
	}

	m.viewport.YOffset = posNum
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) savePos() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Couldn’t locate config dir: log and return
		log.Printf("savePos: cannot locate config dir, pos not saved: %v", err)
		return err
	}

	// Ensure app directory exists: ~/.config/lexo
	appDir := filepath.Join(configDir, "lexo")
	if err := os.MkdirAll(appDir, 0o700); err != nil {
		log.Printf("savePos: cannot create app directory %q: %v", appDir, err)
		return err
	}

	name := sanitize(m.title) + ".pos"
	posPath := filepath.Join(appDir, name)

	// Write the position using secure file permissions
	data := []byte(fmt.Sprint(m.viewport.YOffset))
	if err := os.WriteFile(posPath, data, 0o600); err != nil {
		log.Printf("savePos: cannot write pos file %q: %v", posPath, err)
		return err
	}

	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			err := m.savePos()
			if err != nil {
				log.Fatal("Could not save position", err.Error())
				return m, tea.Quit
			}
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
			m.viewport.SetContent(m.content)
			m.loadYOffset()

			m.ready = true

			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the header.
			m.viewport.YPosition = headerHeight + 1
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

		if useHighPerformanceRenderer {
			// Render (or re-render) the whole viewport. Necessary both to
			// initialize the viewport and when the window is resized.
			//
			// This is needed for high-performance rendering only.
			cmds = append(cmds, viewport.Sync(m.viewport))
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m model) headerView() string {
	title := titleStyle.Render(m.title)
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s <ebook.epub>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	filePath := os.Args[1]

	reader, err := epub.NewEPUBReader(filePath)
	if err != nil {
		fmt.Printf("Error opening EPUB: %v\n", err)
		return
	}

	ep, err := reader.Parse()
	if err != nil {
		fmt.Printf("Error parsing EPUB: %v\n", err)
		return
	}

	content := strings.Join(ep.Contents, "\n\n")

	p := tea.NewProgram(
		model{content: string(content), title: ep.MetaData.Title},
		tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}

func sanitize(title string) string {
	s := strings.TrimSpace(title)
	// Replace everything not a letter or digit with hyphens
	re := regexp.MustCompile(`[^A-Za-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	// Collapse multiple hyphens into one
	s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return strings.ToLower(s)
}
