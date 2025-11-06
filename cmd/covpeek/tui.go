package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Chapati-Systems/covpeek/pkg/models"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// tableModel holds the state for the TUI table
type tableModel struct {
	table        table.Model
	sortCol      int
	sortAsc      bool
	originalRows []table.Row
}

// newTableModel creates a new table model for the TUI
func newTableModel(report *models.CoverageReport) tableModel {
	// Create table columns
	columns := []table.Column{
		{Title: "File", Width: 50},
		{Title: "Total Lines", Width: 12},
		{Title: "Covered Lines", Width: 13},
		{Title: "Coverage %", Width: 11},
	}

	// Create table rows
	var rows []table.Row
	for name, cov := range report.Files {
		rows = append(rows, table.Row{
			name,
			fmt.Sprintf("%d", cov.TotalLines),
			fmt.Sprintf("%d", cov.CoveredLines),
			fmt.Sprintf("%.2f", cov.CoveragePct),
		})
	}

	// Sort rows by coverage descending initially
	sort.Slice(rows, func(i, j int) bool {
		covI, _ := strconv.ParseFloat(rows[i][3], 64)
		covJ, _ := strconv.ParseFloat(rows[j][3], 64)
		return covI > covJ
	})

	// Create table
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	// Set styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	// Update viewport after setting styles
	t.UpdateViewport()

	return tableModel{
		table:        t,
		sortCol:      3,     // Coverage % column
		sortAsc:      false, // descending
		originalRows: make([]table.Row, len(rows)),
	}
}

// Init implements tea.Model
func (m tableModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Handle window resize
		m.table.SetWidth(msg.Width)
		m.table.SetHeight(msg.Height - 4) // Leave room for title/help
		return m, nil
	case tea.MouseMsg:
		// Handle mouse clicks on headers for sorting
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			// Check if click is in header area (roughly top 2 lines)
			if msg.Y <= 2 {
				// Determine which column was clicked based on X position
				colWidths := []int{50, 12, 13, 11} // File, Total Lines, Covered Lines, Coverage %
				x := 0
				for i, width := range colWidths {
					if msg.X >= x && msg.X < x+width {
						m.sortByColumn(i)
						break
					}
					x += width + 1 // +1 for separator
				}
			}
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		// Sorting keys
		case "s":
			// Wait for next key to determine column
			return m, nil
		case "f": // sort by file
			m.sortByColumn(0)
		case "t": // sort by total lines
			m.sortByColumn(1)
		case "c": // sort by covered lines
			m.sortByColumn(2)
		case "p": // sort by coverage %
			m.sortByColumn(3)
		case "r": // reverse sort
			m.sortAsc = !m.sortAsc
			m.sortByColumn(m.sortCol)
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// sortByColumn sorts the table by the specified column
func (m *tableModel) sortByColumn(col int) {
	rows := m.table.Rows()
	m.sortCol = col

	sort.Slice(rows, func(i, j int) bool {
		var a, b string
		if col < len(rows[i]) {
			a = rows[i][col]
		}
		if col < len(rows[j]) {
			b = rows[j][col]
		}

		// For numeric columns, parse as numbers
		if col == 1 || col == 2 || col == 3 { // Total Lines, Covered Lines, Coverage %
			aNum, _ := strconv.ParseFloat(a, 64)
			bNum, _ := strconv.ParseFloat(b, 64)
			if m.sortAsc {
				return aNum < bNum
			}
			return aNum > bNum
		}

		// String comparison for other columns
		if m.sortAsc {
			return a < b
		}
		return a > b
	})

	m.table.SetRows(rows)
}

// View implements tea.Model
func (m tableModel) View() string {
	var b strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Render("Coverage Report")
	b.WriteString(title + "\n\n")

	// Help text
	sortIndicator := "▼"
	if m.sortAsc {
		sortIndicator = "▲"
	}

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(fmt.Sprintf("↑/↓ navigate • click headers to sort • s+r reverse • q quit (sorted by %s %s)",
			[]string{"file", "total", "covered", "coverage"}[m.sortCol], sortIndicator))
	b.WriteString(help + "\n\n")

	// Table
	b.WriteString(m.table.View())

	return b.String()
}
