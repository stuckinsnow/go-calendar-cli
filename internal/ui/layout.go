package ui

// Layout constants describing the minimum usable terminal and pane sizing.
const (
	minWidth       = 44
	minHeight      = 18
	borderWidth    = 1 // rounded border, per side
	paddingWidth   = 1 // panel horizontal padding, per side
	paneGap        = 1 // columns between panes
	minAgendaWidth = 30
	minLegendRows  = 4
	gridFloor      = 35
	gridCeiling    = 63

	// detailFloor/detailCeiling bound the width of the detail side pane, and
	// minAgendaWithDetail is the agenda width below which the third column is
	// abandoned in favour of a centred overlay.
	detailFloor         = 38
	detailCeiling       = 72
	minAgendaWithDetail = 32
)

// computeLayout divides the terminal between the panes.
//
// chromeHeight is everything outside the panes: header plus footer. gridHeight
// is the number of lines the month grid needs. showDetail requests the detail
// side pane, which is granted only when the agenda stays usable.
func computeLayout(width, height, chromeHeight, gridHeight int, showDetail bool) layout {
	if width < minWidth || height < minHeight {
		return layout{tooSmall: true}
	}

	// Rows available for the panes, borders included.
	contentHeight := height - chromeHeight
	if contentHeight < gridHeight+2*borderWidth {
		contentHeight = gridHeight + 2*borderWidth
	}
	paneHeight := contentHeight - 2*borderWidth

	// Preferred grid width, then a narrower one used only if the detail column
	// cannot otherwise fit. Wide terminals therefore keep a full-size grid when
	// the details open; cramped ones trade grid width for it.
	preferred := clamp(width*2/5, gridFloor, gridCeiling)
	narrow := clamp(width/5, gridFloor, gridCeiling)

	if showDetail {
		for _, gridInner := range []int{preferred, narrow} {
			if l, ok := threeColumns(width, gridInner, gridHeight, paneHeight); ok {
				return withLegend(l, contentHeight, paneHeight)
			}
		}
	}

	l, ok := twoColumns(width, preferred, gridHeight, paneHeight)
	if !ok {
		return stackedLayout(width, contentHeight, gridHeight)
	}
	return withLegend(l, contentHeight, paneHeight)
}

// twoColumns is the default arrangement: grid beside agenda.
func twoColumns(width, gridInner, gridHeight, paneHeight int) (layout, bool) {
	monthWidth := gridInner + 2*paddingWidth
	agendaWidth := width - column(monthWidth) - 2*borderWidth

	if agendaWidth-2*paddingWidth < minAgendaWidth {
		return layout{}, false
	}
	return layout{
		month:  pane{Width: monthWidth, Height: gridHeight},
		agenda: pane{Width: agendaWidth, Height: paneHeight},
	}, true
}

// threeColumns adds the detail pane, taken out of the agenda's share.
func threeColumns(width, gridInner, gridHeight, paneHeight int) (layout, bool) {
	l, ok := twoColumns(width, gridInner, gridHeight, paneHeight)
	if !ok {
		return layout{}, false
	}

	detailWidth := clamp(width/3, detailFloor, detailCeiling)
	agendaWidth := l.agenda.Width - column(detailWidth)
	if agendaWidth < minAgendaWithDetail {
		return layout{}, false
	}

	l.showDetail = true
	l.detail = pane{Width: detailWidth, Height: paneHeight}
	l.agenda.Width = agendaWidth
	return l, true
}

// withLegend fills the leftover rows under the grid with the calendar legend so
// that both columns end on the same row; if there is too little space, the grid
// pane simply grows to fill the column.
func withLegend(l layout, contentHeight, paneHeight int) layout {
	legendHeight := contentHeight - l.month.OuterHeight() - 2*borderWidth
	if legendHeight < minLegendRows {
		l.month.Height = paneHeight
		return l
	}

	l.showLegend = true
	l.legend = pane{Width: l.month.Width, Height: legendHeight}
	return l
}

// column is the total width a pane occupies: its content box, its border and
// the gap that follows it.
func column(paneWidth int) int { return paneWidth + 2*borderWidth + paneGap }

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
