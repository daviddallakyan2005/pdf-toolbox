package gui

import (
    "image/color"
    "path/filepath"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/canvas"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/driver/desktop"
    "fyne.io/fyne/v2/layout"
    "fyne.io/fyne/v2/theme"
    "fyne.io/fyne/v2/widget"
)

// SortableList provides drag-to-reorder of a slice of file paths with checkboxes for multi-select
type SortableList struct {
    items    *[]string
    selected *map[int]bool
    box      *fyne.Container
    onChange func()

    // drag state
    dragging bool
    dragFrom int
    dragIdx  int
    dragDY   float32

    // active single-select row (for preview etc.)
    active int
}

func NewSortableList(items *[]string, selected *map[int]bool, onChange func()) *SortableList {
    sl := &SortableList{
        items:    items,
        selected: selected,
        box:      container.NewVBox(),
        onChange: onChange,
    }
    sl.active = -1
    sl.rebuild()
    return sl
}

func (s *SortableList) rebuild() {
    s.box.Objects = nil
    for i, p := range *s.items {
        row := newSortableRow(s, i, filepath.Base(p))
        s.box.Add(row)
    }
    s.box.Refresh()
    s.refreshRowStyles()
}

func (s *SortableList) onRowClick(index int, additive bool) {
    if s.selected == nil { return }
    if additive {
        sel := *s.selected
        sel[index] = !sel[index]
    }
    // Always update active row on click without drag
    s.active = index
    s.refreshRowStyles()
    if s.onChange != nil { s.onChange() }
}

func (s *SortableList) Container() *fyne.Container { return s.box }

func (s *SortableList) move(from, to int) {
    if from == to || from < 0 || to < 0 || from >= len(*s.items) || to >= len(*s.items) {
        return
    }
    items := *s.items
    item := items[from]
    if from < to {
        copy(items[from:to], items[from+1:to+1])
    } else {
        copy(items[to+1:from+1], items[to:from])
    }
    items[to] = item
    // fix selected map reindexing by rebuilding map
    oldSel := *s.selected
    newSel := map[int]bool{}
    for idx := range items {
        if idx == to {
            newSel[idx] = oldSel[from]
            continue
        }
        // map previous indices to new positions
        switch {
        case from < to && idx >= from && idx < to:
            newSel[idx] = oldSel[idx+1]
        case to < from && idx > to && idx <= from:
            newSel[idx] = oldSel[idx-1]
        default:
            newSel[idx] = oldSel[idx]
        }
    }
    *s.selected = newSel
    s.rebuild()
    if s.onChange != nil { s.onChange() }
}

// ---- Row widget ----

type sortableRow struct {
    widget.BaseWidget
    parent   *SortableList
    index    int
    handle   *dragHandle
    check    *widget.Check
    label    *widget.Label
    delBtn   *widget.Button
    bg       *canvas.Rectangle
    dragAccY float32
    lastY    float32
    pressed  bool
    pressMod desktop.Modifier
}

func newSortableRow(parent *SortableList, index int, name string) *sortableRow {
    r := &sortableRow{
        parent: parent,
        index:  index,
        check:  widget.NewCheck("", nil),
        label:  widget.NewLabel(name),
    }
    r.handle = newDragHandle(r)
    r.delBtn = widget.NewButtonWithIcon("", theme.DeleteIcon(), func(){ r.parent.removeAt(r.index) })
    r.check.SetChecked((*parent.selected)[index])
    r.check.OnChanged = func(v bool) {
        (*parent.selected)[index] = v
        if parent.onChange != nil { parent.onChange() }
    }
    r.ExtendBaseWidget(r)
    return r
}

func (r *sortableRow) CreateRenderer() fyne.WidgetRenderer {
    // Background for selection/drag highlighting
    if r.bg == nil {
        r.bg = canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 0})
    }
    content := container.NewHBox(r.handle, r.check, r.label, layout.NewSpacer(), r.delBtn)
    stacked := container.NewStack(r.bg, content)
    return widget.NewSimpleRenderer(stacked)
}

// Implement drag to reorder by accumulating Y movement and swapping when passing a threshold
func (r *sortableRow) Dragged(ev *fyne.DragEvent) {
    if !r.parent.dragging {
        r.parent.beginDrag(r)
    }
    r.parent.onDragDelta(ev.Dragged.DY)
}

func (r *sortableRow) DragEnd() { r.parent.endDrag(); r.dragAccY = 0; r.lastY = 0 }

// hint: show grab cursor when over the handle on desktop
func (r *sortableRow) MouseIn(*desktop.MouseEvent)  {}
func (r *sortableRow) MouseOut()                     {}
func (r *sortableRow) MouseMoved(*desktop.MouseEvent) {}

func (r *sortableRow) MouseDown(ev *desktop.MouseEvent) {
    r.pressed = true
    r.pressMod = ev.Modifier
}

func (r *sortableRow) MouseUp(ev *desktop.MouseEvent) {
    // If a drag was not in progress, treat this as a click/select action
    if r.pressed && !r.parent.dragging {
        additive := (r.pressMod&desktop.ControlModifier) != 0 || (r.pressMod&desktop.SuperModifier) != 0
        r.parent.onRowClick(r.index, additive)
    }
    r.pressed = false
}


func (r *sortableRow) setDraggingStyle(active bool) {
    if r.bg == nil { return }
    if active {
        r.bg.FillColor = color.NRGBA{R: 0x33, G: 0x99, B: 0xff, A: 0x22}
    } else {
        r.bg.FillColor = color.NRGBA{R: 0, G: 0, B: 0, A: 0}
    }
    r.bg.Refresh()
}

// ---- Drag handle widget ----

type dragHandle struct {
    widget.BaseWidget
    row *sortableRow
}

func newDragHandle(row *sortableRow) *dragHandle {
    h := &dragHandle{row: row}
    h.ExtendBaseWidget(h)
    return h
}

func (h *dragHandle) CreateRenderer() fyne.WidgetRenderer {
    // Simple visual grip using text; avoids child widgets so this object receives drag events
    txt := canvas.NewText("≡", nil)
    return widget.NewSimpleRenderer(txt)
}

func (h *dragHandle) Dragged(ev *fyne.DragEvent) {
    h.row.parent.onDragDelta(ev.Dragged.DY)
}
func (h *dragHandle) DragEnd() { h.row.parent.endDrag() }

// Start drag on mouse down to capture the initial offset and set up overlays
func (h *dragHandle) MouseDown(ev *desktop.MouseEvent) {
    h.row.parent.beginDrag(h.row)
}
func (h *dragHandle) MouseUp(*desktop.MouseEvent) {}

// ---- SortableList drag lifecycle ----

func (s *SortableList) beginDrag(r *sortableRow) {
    if s.dragging { return }
    s.dragging = true
    s.dragFrom = r.index
    s.dragIdx = r.index
    s.dragDY = 0
    r.setDraggingStyle(true)
}

// onDragDelta performs neighbor-swaps as the pointer moves, robust to scrolling
func (s *SortableList) onDragDelta(deltaY float32) {
    if !s.dragging { return }
    s.dragDY += deltaY
    // Move downwards
    for s.dragDY > 0 && s.dragIdx < len(*s.items)-1 {
        // Threshold: mid-point of the next row height
        nextObj := s.box.Objects[s.dragIdx+1]
        nextH := nextObj.Size().Height
        if s.dragDY < nextH/2 { break }
        s.dragDY -= nextH
        s.move(s.dragIdx, s.dragIdx+1)
        s.dragIdx++
        s.refreshRowStyles()
    }
    // Move upwards
    for s.dragDY < 0 && s.dragIdx > 0 {
        prevObj := s.box.Objects[s.dragIdx-1]
        prevH := prevObj.Size().Height
        if -s.dragDY < prevH/2 { break }
        s.dragDY += prevH
        s.move(s.dragIdx, s.dragIdx-1)
        s.dragIdx--
        s.refreshRowStyles()
    }
}

func (s *SortableList) endDrag() {
    if !s.dragging { return }
    s.refreshRowStyles()
    // Done
    s.dragging = false
    s.dragFrom = -1
    s.dragIdx = -1
    s.dragDY = 0
}

func (s *SortableList) highlightDragRow() {
    for i, obj := range s.box.Objects {
        if row, ok := obj.(*sortableRow); ok {
            row.index = i
            row.setDraggingStyle(i == s.dragIdx)
        }
    }
}

func (s *SortableList) refreshRowStyles() {
    for i, obj := range s.box.Objects {
        if row, ok := obj.(*sortableRow); ok {
            row.index = i
            // background: dragging row takes precedence, else active selection
            if s.dragging && i == s.dragIdx {
                row.bg.FillColor = color.NRGBA{R: 0x33, G: 0x99, B: 0xff, A: 0x22}
            } else if i == s.active {
                row.bg.FillColor = color.NRGBA{R: 0x33, G: 0x99, B: 0xff, A: 0x44}
            } else {
                row.bg.FillColor = color.NRGBA{R: 0, G: 0, B: 0, A: 0}
            }
            row.bg.Refresh()
            // keep checkbox visuals in sync
            row.check.SetChecked((*s.selected)[i])
        }
    }
}

// ActiveIndex returns the index of the active row selected by click (not checkboxes)
func (s *SortableList) ActiveIndex() int { return s.active }

func (s *SortableList) removeAt(index int) {
    if index < 0 || index >= len(*s.items) { return }
    items := *s.items
    *s.items = append(items[:index], items[index+1:]...)
    // rebuild selected map indices
    oldSel := *s.selected
    newSel := map[int]bool{}
    for i := 0; i < len(*s.items); i++ {
        if i < index {
            newSel[i] = oldSel[i]
        } else {
            newSel[i] = oldSel[i+1]
        }
    }
    *s.selected = newSel
    s.rebuild()
    if s.onChange != nil { s.onChange() }
}


