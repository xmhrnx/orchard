package components

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"orchard/internal/graph"
	"orchard/ui/theme"
)

// Graph geometry, in dp.
const (
	LaneWidth   = 16
	GraphRowH   = 36
	graphStroke = 2
	graphInset  = 10
)

// GraphCellWidth is the width of a graph with the given lane count.
func GraphCellWidth(lanes int) int { return graphInset*2 + max(lanes, 1)*LaneWidth }

// GraphCell draws one row of the commit tree: the lines passing through
// and the commit's node. Lines that change lanes bend in smooth curves,
// so branches read like boughs. head marks the checked-out commit and
// merge draws the node hollow.
func GraphCell(row graph.Row, lanes int, head, merge bool) giode.Element {
	width := GraphCellWidth(lanes)
	return giode.Canvas(giode.Styles{Width: properties.Width(width), Height: GraphRowH}, func(p *giode.Painter) {
		h := p.Height()
		mid := h / 2
		for _, s := range row.Segments {
			c := theme.Lane(s.Color)
			x1, x2 := laneX(s.FromLane), laneX(s.ToLane)
			y1, y2 := float32(0), mid
			if s.Half == graph.Bottom {
				y1, y2 = mid, h
			}
			if x1 == x2 {
				p.Line(x1, y1, x2, y2, graphStroke, c)
				continue
			}
			path := p.Path()
			path.MoveTo(x1, y1)
			bend := (y2 - y1) * 0.85
			path.CubeTo(x1, y1+bend, x2, y2-bend, x2, y2)
			path.Stroke(graphStroke, c)
		}
		x := laneX(row.Lane)
		c := theme.Lane(row.Color)
		switch {
		case head:
			p.Circle(x, mid, 9, giode.Mix(string(c), "theme.surface", 0.25))
			p.Circle(x, mid, 5.5, c)
			p.Circle(x, mid, 2.2, "theme.surface")
		case merge:
			p.Circle(x, mid, 5, c)
			p.Circle(x, mid, 3, "theme.surface")
		default:
			p.Circle(x, mid, 4.5, c)
		}
	})
}

func laneX(lane int) float32 {
	return float32(graphInset + lane*LaneWidth + LaneWidth/2)
}
