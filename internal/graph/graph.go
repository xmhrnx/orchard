// Package graph lays out a commit history as lanes for drawing a tree:
// each commit gets a column, and each row lists the line segments that
// connect it to its neighbours.
package graph

// Half says which half of a row a segment spans.
type Half int

// Segment halves: Top runs from the row's top edge to its middle, Bottom
// from the middle to the bottom edge.
const (
	Top Half = iota
	Bottom
)

// Segment is a line inside one row, from FromLane to ToLane.
type Segment struct {
	FromLane int
	ToLane   int
	Half     Half
	Color    int
}

// Row is the drawing of one commit.
type Row struct {
	Lane     int
	Color    int
	Segments []Segment
}

// Node is the minimal commit shape the layout needs.
type Node struct {
	Hash    string
	Parents []string
}

// Layout is the result of laying out a history.
type Layout struct {
	Rows  []Row
	Lanes int
}

// lane is a column waiting for the commit Hash to appear.
type lane struct {
	hash  string
	color int
}

// Compute lays out nodes given newest first, as git log prints them.
func Compute(nodes []Node) Layout {
	var (
		lanes     []lane
		nextColor int
		out       = Layout{Rows: make([]Row, 0, len(nodes))}
	)
	newColor := func() int { c := nextColor; nextColor++; return c }

	// Parents outside the loaded range (a shallow clone, or a history
	// cut at a limit) never arrive; a lane waiting for one would stay
	// open forever and push every later lane sideways.
	present := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		present[n.Hash] = true
	}

	for _, n := range nodes {
		parents := n.Parents[:0:0]
		for _, p := range n.Parents {
			if present[p] {
				parents = append(parents, p)
			}
		}
		idx := indexOf(lanes, n.Hash)
		if idx < 0 {
			idx = freeLane(&lanes, -1)
			lanes[idx] = lane{hash: n.Hash, color: newColor()}
		}
		row := Row{Lane: idx, Color: lanes[idx].color}

		for j, l := range lanes {
			switch {
			case l.hash == "":
			case l.hash == n.Hash:
				row.Segments = append(row.Segments, Segment{FromLane: j, ToLane: idx, Half: Top, Color: l.color})
				if j != idx {
					lanes[j] = lane{}
				}
			default:
				row.Segments = append(row.Segments, Segment{FromLane: j, ToLane: j, Half: Top, Color: l.color})
			}
		}

		passing := make([]bool, len(lanes))
		for j, l := range lanes {
			passing[j] = l.hash != "" && j != idx
		}

		if len(parents) == 0 {
			lanes[idx] = lane{}
		} else {
			lanes[idx].hash = parents[0]
			row.Segments = append(row.Segments, Segment{FromLane: idx, ToLane: idx, Half: Bottom, Color: lanes[idx].color})
		}
		for _, p := range parents[min(1, len(parents)):] {
			k := indexOf(lanes, p)
			if k < 0 {
				k = freeLane(&lanes, idx)
				lanes[k] = lane{hash: p, color: newColor()}
			}
			row.Segments = append(row.Segments, Segment{FromLane: idx, ToLane: k, Half: Bottom, Color: lanes[k].color})
		}
		for j, pass := range passing {
			if pass {
				row.Segments = append(row.Segments, Segment{FromLane: j, ToLane: j, Half: Bottom, Color: lanes[j].color})
			}
		}

		lanes = trimLanes(lanes)
		out.Lanes = max(out.Lanes, len(lanes), idx+1)
		out.Rows = append(out.Rows, row)
	}
	return out
}

func indexOf(lanes []lane, hash string) int {
	for i, l := range lanes {
		if l.hash == hash {
			return i
		}
	}
	return -1
}

// freeLane returns the first empty lane other than skip, growing the
// slice when every lane is taken.
func freeLane(lanes *[]lane, skip int) int {
	for i, l := range *lanes {
		if l.hash == "" && i != skip {
			return i
		}
	}
	*lanes = append(*lanes, lane{})
	return len(*lanes) - 1
}

func trimLanes(lanes []lane) []lane {
	for len(lanes) > 0 && lanes[len(lanes)-1].hash == "" {
		lanes = lanes[:len(lanes)-1]
	}
	return lanes
}
