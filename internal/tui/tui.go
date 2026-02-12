package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/stackgen-cli/dataclean/internal/models"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true).Foreground(lipgloss.Color("39"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	warningStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	successStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("40"))
)

// VolumeItem represents a volume in the TUI list
type VolumeItem struct {
	Volume   models.Volume
	Selected bool
}

func (i VolumeItem) FilterValue() string { return i.Volume.Name }
func (i VolumeItem) Title() string       { return i.Volume.Name }
func (i VolumeItem) Description() string {
	_, icon := models.GetDatastoreInfo(i.Volume.DatastoreType)
	return fmt.Sprintf("%s %s", icon, i.Volume.DatastoreType)
}

// SnapshotItem represents a snapshot in the TUI list
type SnapshotItem struct {
	Snapshot models.Snapshot
}

func (i SnapshotItem) FilterValue() string { return i.Snapshot.Name }
func (i SnapshotItem) Title() string       { return i.Snapshot.Name }
func (i SnapshotItem) Description() string {
	return fmt.Sprintf("%s | %s | %d volumes",
		i.Snapshot.Timestamp.Format("2006-01-02 15:04"),
		i.Snapshot.SizeHuman,
		len(i.Snapshot.Volumes))
}

// Model represents the TUI application state
type Model struct {
	list       list.Model
	mode       Mode
	volumes    []models.Volume
	snapshots  []models.Snapshot
	selected   []models.Volume
	err        error
	quitting   bool
	confirmed  bool
	width      int
	height     int
}

// Mode represents the current TUI mode
type Mode int

const (
	ModeSelectVolumes Mode = iota
	ModeSelectSnapshot
	ModeConfirmDestructive
)

// NewVolumeSelector creates a TUI for selecting volumes
func NewVolumeSelector(volumes []models.Volume) Model {
	items := make([]list.Item, len(volumes))
	for i, v := range volumes {
		items[i] = VolumeItem{Volume: v}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Volumes"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	return Model{
		list:    l,
		mode:    ModeSelectVolumes,
		volumes: volumes,
	}
}

// NewSnapshotSelector creates a TUI for selecting a snapshot
func NewSnapshotSelector(snapshots []models.Snapshot) Model {
	items := make([]list.Item, len(snapshots))
	for i, s := range snapshots {
		items[i] = SnapshotItem{Snapshot: s}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Snapshot"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	return Model{
		list:      l,
		mode:      ModeSelectSnapshot,
		snapshots: snapshots,
	}
}

// Init implements bubbletea.Model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements bubbletea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.mode == ModeSelectVolumes {
				// Get selected item
				if item, ok := m.list.SelectedItem().(VolumeItem); ok {
					m.selected = append(m.selected, item.Volume)
				}
				m.quitting = true
				m.confirmed = true
				return m, tea.Quit
			}
			if m.mode == ModeSelectSnapshot {
				m.quitting = true
				m.confirmed = true
				return m, tea.Quit
			}
		case " ":
			// Toggle selection for multi-select
			if m.mode == ModeSelectVolumes {
				if item, ok := m.list.SelectedItem().(VolumeItem); ok {
					item.Selected = !item.Selected
					// Update item in list
					items := m.list.Items()
					items[m.list.Index()] = item
					m.list.SetItems(items)
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements bubbletea.Model
func (m Model) View() string {
	if m.quitting {
		return ""
	}
	return m.list.View()
}

// SelectedVolumes returns the selected volumes
func (m Model) SelectedVolumes() []models.Volume {
	return m.selected
}

// SelectedSnapshot returns the selected snapshot
func (m Model) SelectedSnapshot() *models.Snapshot {
	if item, ok := m.list.SelectedItem().(SnapshotItem); ok {
		return &item.Snapshot
	}
	return nil
}

// Confirmed returns whether the user confirmed the selection
func (m Model) Confirmed() bool {
	return m.confirmed
}

// RunVolumeSelector runs the volume selection TUI
func RunVolumeSelector(volumes []models.Volume) ([]models.Volume, error) {
	m := NewVolumeSelector(volumes)
	p := tea.NewProgram(m, tea.WithAltScreen())
	
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm := finalModel.(Model)
	if !fm.Confirmed() {
		return nil, fmt.Errorf("cancelled")
	}

	return fm.SelectedVolumes(), nil
}

// RunSnapshotSelector runs the snapshot selection TUI
func RunSnapshotSelector(snapshots []models.Snapshot) (*models.Snapshot, error) {
	m := NewSnapshotSelector(snapshots)
	p := tea.NewProgram(m, tea.WithAltScreen())
	
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm := finalModel.(Model)
	if !fm.Confirmed() {
		return nil, fmt.Errorf("cancelled")
	}

	return fm.SelectedSnapshot(), nil
}

// ConfirmDestructive shows a confirmation prompt for destructive operations
func ConfirmDestructive(message string) (bool, error) {
	fmt.Println(warningStyle.Render("⚠️  " + message))
	fmt.Print("Type 'yes' to confirm: ")
	
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false, err
	}
	
	return response == "yes", nil
}
