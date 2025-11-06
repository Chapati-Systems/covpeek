package main

import (
	"testing"

	"git.kernel.fun/chapati.systems/covpeek/pkg/models"
	tea "github.com/charmbracelet/bubbletea"
)

func TestTableModel_Init(t *testing.T) {
	report := models.NewCoverageReport()
	model := newTableModel(report)
	cmd := model.Init()
	if cmd != nil {
		t.Error("Expected nil command from Init")
	}
}

func TestTableModel_View(t *testing.T) {
	report := models.NewCoverageReport()
	model := newTableModel(report)
	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestTableModel_Update_Quit(t *testing.T) {
	report := models.NewCoverageReport()
	model := newTableModel(report)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newModel, cmd := model.Update(msg)
	if cmd == nil {
		t.Error("Expected quit command")
	}
	if newModel == nil {
		t.Error("Expected model")
	}
}

func TestTableModel_Update_UnknownKey(t *testing.T) {
	report := models.NewCoverageReport()
	model := newTableModel(report)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	newModel, cmd := model.Update(msg)
	if cmd != nil {
		t.Error("Expected no command for unknown key")
	}
	if newModel == nil {
		t.Error("Expected model")
	}
}

func TestTableModel_Update_Enter(t *testing.T) {
	report := models.NewCoverageReport()
	model := newTableModel(report)
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := model.Update(msg)
	if cmd == nil {
		t.Error("Expected quit command for enter")
	}
	if newModel == nil {
		t.Error("Expected model")
	}
}

func TestTableModel_Update_Up(t *testing.T) {
	report := models.NewCoverageReport()
	model := newTableModel(report)
	msg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, cmd := model.Update(msg)
	if cmd != nil {
		t.Error("Expected no command for up")
	}
	if newModel == nil {
		t.Error("Expected model")
	}
}

func TestTableModel_Update_Down(t *testing.T) {
	report := models.NewCoverageReport()
	model := newTableModel(report)
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, cmd := model.Update(msg)
	if cmd != nil {
		t.Error("Expected no command for down")
	}
	if newModel == nil {
		t.Error("Expected model")
	}
}

func TestTableModel_Update_Sort(t *testing.T) {
	report := models.NewCoverageReport()
	model := newTableModel(report)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	newModel, cmd := model.Update(msg)
	if cmd != nil {
		t.Error("Expected no command for sort")
	}
	if newModel == nil {
		t.Error("Expected model")
	}
}

func TestTableModel_Update_WindowSize(t *testing.T) {
	report := models.NewCoverageReport()
	model := newTableModel(report)
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, cmd := model.Update(msg)
	if cmd != nil {
		t.Error("Expected no command for window size")
	}
	if newModel == nil {
		t.Error("Expected model")
	}
}

func TestTableModel_Update_SortByFile(t *testing.T) {
	report := models.NewCoverageReport()
	model := newTableModel(report)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	newModel, cmd := model.Update(msg)
	if cmd != nil {
		t.Error("Expected no command for sort by file")
	}
	if newModel == nil {
		t.Error("Expected model")
	}
}

func TestTableModel_Update_SortByTotal(t *testing.T) {
	report := models.NewCoverageReport()
	model := newTableModel(report)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	newModel, cmd := model.Update(msg)
	if cmd != nil {
		t.Error("Expected no command for sort by total")
	}
	if newModel == nil {
		t.Error("Expected model")
	}
}

func TestTableModel_Update_SortByCovered(t *testing.T) {
	report := models.NewCoverageReport()
	model := newTableModel(report)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	newModel, cmd := model.Update(msg)
	if cmd != nil {
		t.Error("Expected no command for sort by covered")
	}
	if newModel == nil {
		t.Error("Expected model")
	}
}

func TestTableModel_Update_SortByPercent(t *testing.T) {
	report := models.NewCoverageReport()
	model := newTableModel(report)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	newModel, cmd := model.Update(msg)
	if cmd != nil {
		t.Error("Expected no command for sort by percent")
	}
	if newModel == nil {
		t.Error("Expected model")
	}
}

func TestTableModel_Update_ReverseSort(t *testing.T) {
	report := models.NewCoverageReport()
	model := newTableModel(report)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	newModel, cmd := model.Update(msg)
	if cmd != nil {
		t.Error("Expected no command for reverse sort")
	}
	if newModel == nil {
		t.Error("Expected model")
	}
}

func TestTableModel_Update_Esc(t *testing.T) {
	report := models.NewCoverageReport()
	model := newTableModel(report)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, cmd := model.Update(msg)
	if cmd != nil {
		t.Error("Expected no command for esc")
	}
	if newModel == nil {
		t.Error("Expected model")
	}
}

func TestTableModel_Update_CtrlC(t *testing.T) {
	report := models.NewCoverageReport()
	model := newTableModel(report)
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	newModel, cmd := model.Update(msg)
	if cmd == nil {
		t.Error("Expected quit command for ctrl+c")
	}
	if newModel == nil {
		t.Error("Expected model")
	}
}

func TestTableModel_Update_MouseClick(t *testing.T) {
	report := models.NewCoverageReport()
	model := newTableModel(report)
	msg := tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, Y: 1, X: 10}
	newModel, cmd := model.Update(msg)
	if cmd != nil {
		t.Error("Expected no command for mouse click")
	}
	if newModel == nil {
		t.Error("Expected model")
	}
}
