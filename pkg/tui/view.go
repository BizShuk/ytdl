package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#1E3A8A"))

	bannerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E2E8F0")).
			Background(lipgloss.Color("#0F172A"))

	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#94A3B8"))

	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#2563EB"))

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#38BDF8"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748B"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8")).
			Background(lipgloss.Color("#0F172A"))

	noticeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FACC15")).
			Background(lipgloss.Color("#0F172A"))
)

// View renders the TUI layout to fit exact width × height dimensions.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	w := m.width
	h := m.height

	// 1. Header (1 line)
	headerLeft := fmt.Sprintf(" ytdl monitor  %d 檔案 · %s", len(m.files), FormatBytes(m.totalBytes))
	headerRight := m.dir + " "
	spaceCount := w - stringWidth(headerLeft) - stringWidth(headerRight)
	if spaceCount < 1 {
		// Truncate directory path if terminal is narrow
		avail := w - stringWidth(headerLeft) - 2
		headerRight = truncate(headerRight, max(0, avail), "…") + " "
		spaceCount = max(1, w-stringWidth(headerLeft)-stringWidth(headerRight))
	}
	headerRaw := headerLeft + strings.Repeat(" ", spaceCount) + headerRight
	headerLine := headerStyle.Render(headerRaw)

	// 2. Banner (1 line)
	var bannerRaw string
	if m.searchMode {
		bannerRaw = fmt.Sprintf(" / %s_  (Enter 搜尋, Esc 取消)", m.searchQuery)
	} else {
		tabAll := "All"
		tabMP3 := "mp3"
		tabMP4 := "mp4"
		switch m.filterType {
		case "all":
			tabAll = activeTabStyle.Render("[All]")
		case "mp3":
			tabMP3 = activeTabStyle.Render("[mp3]")
		case "mp4":
			tabMP4 = activeTabStyle.Render("[mp4]")
		}
		sortLabel := m.sortMode.Label()
		bannerRaw = fmt.Sprintf(" 格式: %s  %s  %s (tab 切換) │ 排序: %s (s 切換)", tabAll, tabMP3, tabMP4, sortLabel)
		if m.searchQuery != "" {
			bannerRaw += fmt.Sprintf(" │ 搜尋: %q", m.searchQuery)
		}
	}
	bannerRaw = padRight(bannerRaw, w)
	bannerLine := bannerStyle.Render(bannerRaw)

	// 3. Footer (1 line)
	var footerRaw string
	if m.notice != "" {
		footerRaw = noticeStyle.Render(" " + m.notice)
	} else {
		footerRaw = " ↑↓/jk 移動 │ enter/o 開啟 │ / 搜尋 │ tab 格式 │ s 排序 │ r 重新整理 │ q 離開"
	}
	footerRaw = padRight(footerRaw, w)
	footerLine := footerStyle.Render(footerRaw)

	// 4. Body (h - 3 lines)
	bodyHeight := h - 3
	if bodyHeight < 1 {
		return lipgloss.JoinVertical(lipgloss.Left, headerLine, footerLine)
	}

	// Column widths
	const (
		colPrefix = 3  // " > " or "   "
		colExt    = 6  // "mp3"
		colSize   = 10 // "12.3 MB"
		colTime   = 17 // "2026-09-13 16:42"
		colGaps   = 3  // spaces between columns
	)
	fixedCols := colPrefix + colExt + colSize + colTime + colGaps
	colName := max(15, w-fixedCols)

	// Table column headers
	thPrefix := strings.Repeat(" ", colPrefix)
	thName := padRight("名稱", colName)
	thExt := padRight("格式", colExt)
	thSize := padRight("大小", colSize)
	thTime := padRight("下載時間", colTime)
	tableHeaderRaw := thPrefix + thName + " " + thExt + " " + thSize + " " + thTime
	tableHeaderLine := tableHeaderStyle.Render(padRight(tableHeaderRaw, w))

	var rows []string
	rows = append(rows, tableHeaderLine)

	dataSlots := bodyHeight - 1
	totalItems := len(m.filtered)

	if totalItems == 0 {
		var emptyMsg string
		if len(m.files) == 0 {
			emptyMsg = fmt.Sprintf("   (目錄尚無下載檔案: %s)", m.dir)
		} else {
			emptyMsg = "   (無符合條件之檔案)"
		}
		rows = append(rows, dimStyle.Render(padRight(emptyMsg, w)))
		for len(rows) < bodyHeight {
			rows = append(rows, strings.Repeat(" ", w))
		}
	} else {
		// Cursor-centered window calculation per tui.md
		visible := dataSlots
		start := max(0, min(m.cursor-visible/2, totalItems-visible))
		end := min(start+visible, totalItems)

		for i := start; i < end; i++ {
			item := m.filtered[i]
			isSel := i == m.cursor

			prefix := "   "
			if isSel {
				prefix = " > "
			}

			nameCell := padRight(truncate(item.Name, colName, "…"), colName)
			extCell := padRight(item.Ext, colExt)
			sizeCell := padRight(FormatBytes(item.Size), colSize)
			timeCell := padRight(FormatTime(item.ModTime), colTime)

			rowRaw := prefix + nameCell + " " + extCell + " " + sizeCell + " " + timeCell
			rowRaw = padRight(rowRaw, w)

			if isSel {
				rows = append(rows, selectedRowStyle.Render(rowRaw))
			} else {
				rows = append(rows, rowRaw)
			}
		}

		// Fill blank rows to keep exact height budget
		for len(rows) < bodyHeight {
			rows = append(rows, strings.Repeat(" ", w))
		}
	}

	bodyContent := strings.Join(rows, "\n")
	return lipgloss.JoinVertical(lipgloss.Left, headerLine, bannerLine, bodyContent, footerLine)
}
