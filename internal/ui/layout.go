package ui

// Layout constants describing the minimum usable terminal and pane sizing.
const (
	minWidth       = 44
	minHeight      = 18
	borderWidth    = 1 // rounded border, per side
	paddingWidth   = 1 // panel horizontal padding, per side
	paneGap        = 1 // columns between the two panes
	minAgendaWidth = 30
	minLegendRows  = 4
	gridFloor      = 35
	gridCeiling    = 63
)

// computeLayout divides the terminal between the panes.
//
// chromeHeight is everything outside the panes: header, status bar and help.
// gridHeight is the number of lines the month grid needs.
func computeLayout(width, height, chromeHeight, gridHeight int) layout {
	var l layout
	if width < minWidth || height < minHeight {
		l.tooSmall = true
		return l
	}

	// Rows available for the panes, borders included.
	contentHeight := height - chromeHeight
	if contentHeight < gridHeight+2*borderWidth {
		contentHeight = gridHeight + 2*borderWidth
	}

	gridInner := clamp(width*2/5, gridFloor, gridCeiling)
	monthWidth := gridInner + 2*paddingWidth
	agendaWidth := width - (monthWidth + 2*borderWidth) - paneGap - 2*borderWidth

	if agendaWidth-2*paddingWidth < minAgendaWidth {
		return stackedLayout(width, contentHeight, gridHeight)
	}

	l.month = pane{Width: monthWidth, Height: gridHeight}
	l.agenda = pane{Width: agendaWidth, Height: contentHeight - 2*borderWidth}

	// Whatever is left in the left column becomes the legend, so both columns
	// end on the same row.
	legendHeight := contentHeight - l.month.OuterHeight() - 2*borderWidth
	if legendHeight >= minLegendRows {
		l.showLegend = true
		l.legend = pane{Width: monthWidth, Height: legendHeight}
	} else {
		l.month.Height = contentHeight - 2*borderWidth
	}
	return l
}

// stackedLayout puts the grid above the agenda for narrow terminals.
func stackedLayout(width, contentHeight, gridHeight int) layout {
	paneWidth := width - 2*borderWidth

	agendaHeight := contentHeight - (gridHeight + 2*borderWidth) - 2*borderWidth
	if agendaHeight < 3 {
		agendaHeight = 3
	}

	return layout{
		stacked: true,
		month:   pane{Width: paneWidth, Height: gridHeight},
		agenda:  pane{Width: paneWidth, Height: agendaHeight},
	}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
