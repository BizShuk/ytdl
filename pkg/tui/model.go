package tui

import (
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type SortMode int

const (
	SortDateDesc SortMode = iota
	SortDateAsc
	SortSizeDesc
	SortSizeAsc
	SortNameAsc
	SortNameDesc
	sortModeCount
)

func (s SortMode) Label() string {
	switch s {
	case SortDateDesc:
		return "時間 ↓"
	case SortDateAsc:
		return "時間 ↑"
	case SortSizeDesc:
		return "大小 ↓"
	case SortSizeAsc:
		return "大小 ↑"
	case SortNameAsc:
		return "名稱 ↑"
	case SortNameDesc:
		return "名稱 ↓"
	default:
		return "時間 ↓"
	}
}

type tickMsg time.Time

type refreshMsg struct {
	files []FileItem
	err   error
}

// Model maintains the state of the ytdl monitor TUI.
type Model struct {
	dir          string
	files        []FileItem
	filtered     []FileItem
	cursor       int
	width        int
	height       int
	searchMode   bool
	searchQuery  string
	filterType   string // "all", "mp3", "mp4"
	sortMode     SortMode
	notice       string
	noticeExpiry time.Time
	totalBytes   int64
}

// NewModel creates an initial model seeded with default dimensions per tui.md.
func NewModel(dir string) Model {
	m := Model{
		dir:        dir,
		width:      100,
		height:     30,
		filterType: "all",
		sortMode:   SortDateDesc,
	}
	return m
}

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func reloadCmd(dir string) tea.Cmd {
	return func() tea.Msg {
		files, err := LoadFiles(dir)
		return refreshMsg{files: files, err: err}
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(reloadCmd(m.dir), tickCmd())
}

func (m *Model) applyFilterAndSort() {
	var matched []FileItem
	var total int64
	for _, f := range m.files {
		total += f.Size
		// Filter by format
		if m.filterType != "all" && !strings.EqualFold(f.Ext, m.filterType) {
			continue
		}
		// Filter by subsequence search query
		if m.searchQuery != "" && !SubsequenceMatch(m.searchQuery, f.Name) {
			continue
		}
		matched = append(matched, f)
	}
	m.totalBytes = total

	// Sort
	switch m.sortMode {
	case SortDateDesc:
		sort.SliceStable(matched, func(i, j int) bool {
			return matched[i].ModTime.After(matched[j].ModTime)
		})
	case SortDateAsc:
		sort.SliceStable(matched, func(i, j int) bool {
			return matched[i].ModTime.Before(matched[j].ModTime)
		})
	case SortSizeDesc:
		sort.SliceStable(matched, func(i, j int) bool {
			return matched[i].Size > matched[j].Size
		})
	case SortSizeAsc:
		sort.SliceStable(matched, func(i, j int) bool {
			return matched[i].Size < matched[j].Size
		})
	case SortNameAsc:
		sort.SliceStable(matched, func(i, j int) bool {
			return strings.ToLower(matched[i].Name) < strings.ToLower(matched[j].Name)
		})
	case SortNameDesc:
		sort.SliceStable(matched, func(i, j int) bool {
			return strings.ToLower(matched[i].Name) > strings.ToLower(matched[j].Name)
		})
	}

	m.filtered = matched
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		if time.Now().After(m.noticeExpiry) {
			m.notice = ""
		}
		return m, tea.Batch(reloadCmd(m.dir), tickCmd())

	case refreshMsg:
		if msg.err == nil {
			var selectedPath string
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				selectedPath = m.filtered[m.cursor].Path
			}
			m.files = msg.files
			m.applyFilterAndSort()

			// Restore selection across refreshes
			if selectedPath != "" {
				for i, f := range m.filtered {
					if f.Path == selectedPath {
						m.cursor = i
						break
					}
				}
			}
		}
		return m, nil

	case tea.KeyMsg:
		// Search mode typing gate
		if m.searchMode {
			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyEsc:
				m.searchMode = false
				m.searchQuery = ""
				m.applyFilterAndSort()
				return m, nil
			case tea.KeyEnter:
				m.searchMode = false
				return m, nil
			case tea.KeyBackspace:
				if len(m.searchQuery) > 0 {
					runes := []rune(m.searchQuery)
					m.searchQuery = string(runes[:len(runes)-1])
					m.applyFilterAndSort()
				}
				return m, nil
			case tea.KeyRunes, tea.KeySpace:
				m.searchQuery += msg.String()
				m.applyFilterAndSort()
				return m, nil
			default:
				return m, nil
			}
		}

		// Normal mode keys
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "/":
			m.searchMode = true
			m.searchQuery = ""
			m.applyFilterAndSort()
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
			return m, nil

		case "home", "g":
			m.cursor = 0
			return m, nil

		case "end", "G":
			if len(m.filtered) > 0 {
				m.cursor = len(m.filtered) - 1
			}
			return m, nil

		case "tab", "t":
			switch m.filterType {
			case "all":
				m.filterType = "mp3"
			case "mp3":
				m.filterType = "mp4"
			default:
				m.filterType = "all"
			}
			m.applyFilterAndSort()
			return m, nil

		case "s":
			m.sortMode = (m.sortMode + 1) % sortModeCount
			m.applyFilterAndSort()
			return m, nil

		case "r":
			return m, reloadCmd(m.dir)

		case "enter", "o":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				target := m.filtered[m.cursor]
				if err := OpenFile(target.Path); err != nil {
					m.notice = "開啟失敗: " + err.Error()
				} else {
					m.notice = "已開啟: " + target.Name
				}
				m.noticeExpiry = time.Now().Add(3 * time.Second)
			}
			return m, nil
		}
	}

	return m, nil
}
