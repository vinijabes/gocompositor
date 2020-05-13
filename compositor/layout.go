package compositor

import (
	"fmt"
)

//LayoutSlot ...
type LayoutSlot struct {
	posx int64
	posy int64

	sizex int64
	sizey int64

	borderTop    int64
	borderRight  int64
	borderBottom int64
	borderLeft   int64

	group string
}

//LayoutRule ...
type LayoutRule struct {
	slots []*LayoutSlot
}

//Layout ...
type Layout struct {
	width  int64
	height int64

	rules map[int]*LayoutRule
}

//NewLayout ...
func NewLayout(width int, height int) *Layout {
	layout := &Layout{
		width:  int64(width),
		height: int64(height),
	}

	layout.rules = make(map[int]*LayoutRule)

	return layout
}

//NewLayoutRule ...
func NewLayoutRule() *LayoutRule {
	return &LayoutRule{}
}

//NewLayoutSlot ...
func NewLayoutSlot(posx int, posy int, sizex int, sizey int) *LayoutSlot {
	slot := &LayoutSlot{
		posx:  int64(posx),
		posy:  int64(posy),
		sizex: int64(sizex),
		sizey: int64(sizey),
		group: "default",
	}

	return slot
}

//NewLayoutSlotWithBorders ...
func NewLayoutSlotWithBorders(posx int, posy int, sizex int, sizey int, borderTop int, borderRight int, borderBottom int, borderLeft int) *LayoutSlot {
	slot := &LayoutSlot{
		posx:         int64(posx),
		posy:         int64(posy),
		sizex:        int64(sizex),
		sizey:        int64(sizey),
		borderTop:    int64(borderTop),
		borderRight:  int64(borderRight),
		borderBottom: int64(borderBottom),
		borderLeft:   int64(borderLeft),
		group:        "default",
	}

	return slot
}

//NewLayoutSlotWithSymetricBorders ...
func NewLayoutSlotWithSymetricBorders(posx int, posy int, sizex int, sizey int, horizontalBorder int, verticalBorder int) *LayoutSlot {
	slot := &LayoutSlot{
		posx:         int64(posx),
		posy:         int64(posy),
		sizex:        int64(sizex),
		sizey:        int64(sizey),
		borderTop:    int64(verticalBorder),
		borderRight:  int64(horizontalBorder),
		borderBottom: int64(verticalBorder),
		borderLeft:   int64(horizontalBorder),
		group:        "default",
	}

	return slot
}

//AddRule ...
func (l *Layout) AddRule(r *LayoutRule, videoAmount int) {
	l.rules[videoAmount] = r
}

//ApplyLayout ...
func (l *Layout) ApplyLayout(videos Videos) error {
	if rule, ok := l.rules[len(videos)]; ok {
		for i := 0; i < len(videos); i++ {
			rule.slots[i].applyLayout(videos[i])
		}
	} else {
		return fmt.Errorf("No matching rule for %d input sources", len(videos))
	}

	return nil
}

func (l *LayoutSlot) applyLayout(video Video) {
	video.SetPos(l.posx, l.posy)
	video.SetSize(l.sizex, l.sizey)
	video.SetBorder("left", -l.borderLeft)
	video.SetBorder("right", -l.borderRight)
	video.SetBorder("top", -l.borderTop)
	video.SetBorder("bottom", -l.borderBottom)
}

//AddSlot ...
func (lr *LayoutRule) AddSlot(ls *LayoutSlot) {
	lr.slots = append(lr.slots, ls)
}
