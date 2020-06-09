package compositor

import (
	"fmt"

	"github.com/vinijabes/gocompositor/pkg/compositor/element"
)

//LayoutSlot ...
type LayoutSlot struct {
	posx int
	posy int

	sizex int
	sizey int

	borderTop    int
	borderRight  int
	borderBottom int
	borderLeft   int

	group string
}

//LayoutRule ...
type LayoutRule struct {
	slots []*LayoutSlot
}

//Layout ...
type Layout struct {
	width  int
	height int

	rules map[int]*LayoutRule
}

//NewLayout ...
func NewLayout(width int, height int) *Layout {
	layout := &Layout{
		width:  int(width),
		height: int(height),
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
		posx:  posx,
		posy:  posy,
		sizex: sizex,
		sizey: sizey,
		group: "default",
	}

	return slot
}

//NewLayoutSlotWithBorders ...
func NewLayoutSlotWithBorders(posx int, posy int, sizex int, sizey int, borderTop int, borderRight int, borderBottom int, borderLeft int) *LayoutSlot {
	slot := &LayoutSlot{
		posx:         posx,
		posy:         posy,
		sizex:        sizex,
		sizey:        sizey,
		borderTop:    borderTop,
		borderRight:  borderRight,
		borderBottom: borderBottom,
		borderLeft:   borderLeft,
		group:        "default",
	}

	return slot
}

//NewLayoutSlotWithSymetricBorders ...
func NewLayoutSlotWithSymetricBorders(posx int, posy int, sizex int, sizey int, horizontalBorder int, verticalBorder int) *LayoutSlot {
	slot := &LayoutSlot{
		posx:         posx,
		posy:         posy,
		sizex:        sizex,
		sizey:        sizey,
		borderTop:    verticalBorder,
		borderRight:  horizontalBorder,
		borderBottom: verticalBorder,
		borderLeft:   horizontalBorder,
		group:        "default",
	}

	return slot
}

//AddRule ...
func (l *Layout) AddRule(r *LayoutRule, videoAmount int) {
	l.rules[videoAmount] = r
}

//ApplyLayout ...
func (l *Layout) ApplyLayout(videos element.Videos) error {
	if rule, ok := l.rules[len(videos)]; ok {
		for i := 0; i < len(videos); i++ {
			rule.slots[i].applyLayout(videos[i])
		}
	} else {
		return fmt.Errorf("No matching rule for %d input sources", len(videos))
	}

	return nil
}

func (l *LayoutSlot) applyLayout(video element.Video) {
	video.SetPos(l.posx, l.posy)
	video.SetSize(l.sizex, l.sizey)
	video.SetBorder(element.VideoBorderLeft, -l.borderLeft)
	video.SetBorder(element.VideoBorderRight, -l.borderRight)
	video.SetBorder(element.VideoBorderTop, -l.borderTop)
	video.SetBorder(element.VideoBorderBottom, -l.borderBottom)
}

//AddSlot ...
func (lr *LayoutRule) AddSlot(ls *LayoutSlot) {
	lr.slots = append(lr.slots, ls)
}
