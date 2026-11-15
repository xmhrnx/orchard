package theme

import "github.com/xmhrnx/giode"

// laneColors colour the commit graph's lanes; they read on both the
// dark and light canvas.
var laneColors = []giode.Color{
	"#23B183", "#F5B54A", "#60A5FA", "#F472B6", "#A78BFA",
	"#2DD4BF", "#FB923C", "#A3E635", "#F87171", "#38BDF8",
}

// Lane returns the colour of graph lane colour index i.
func Lane(i int) giode.Color { return laneColors[i%len(laneColors)] }
