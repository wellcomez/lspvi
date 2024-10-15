package mainui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *app) SetFocus(p tview.Primitive) {
}

type app struct {
	root tview.Primitive
	mouseCapture            func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction)
	mouseCapturingPrimitive tview.Primitive  // A Primitive returned by a MouseHandler which will capture future mouse events.
	lastMouseX, lastMouseY  int              // The last position of the mouse.
	mouseDownX, mouseDownY  int              // The position of the mouse when its button was last pressed.
	lastMouseClick          time.Time        // The time when a mouse button was last clicked.
	lastMouseButtons        tcell.ButtonMask // The
}

// fireMouseActions analyzes the provided mouse event, derives mouse actions
// from it and then forwards them to the corresponding primitives.
func (a *app) fireMouseActions(event *tcell.EventMouse) (consumed, isMouseDownAction bool) {
	// We want to relay follow-up events to the same target primitive.
	var targetPrimitive tview.Primitive

	// Helper function to fire a mouse action.
	fire := func(action tview.MouseAction) {
		switch action {
		case tview.MouseLeftDown, tview.MouseMiddleDown, tview.MouseRightDown:
			isMouseDownAction = true
		}

		// Intercept event.
		if a.mouseCapture != nil {
			event, action = a.mouseCapture(event, action)
			if event == nil {
				consumed = true
				return // Don't forward event.
			}
		}

		// Determine the target primitive.
		var primitive, capturingPrimitive tview.Primitive
		if a.mouseCapturingPrimitive != nil {
			primitive = a.mouseCapturingPrimitive
			targetPrimitive = a.mouseCapturingPrimitive
		} else if targetPrimitive != nil {
			primitive = targetPrimitive
		} else {
			primitive = a.root
		}
		if primitive != nil {
			if handler := primitive.MouseHandler(); handler != nil {
				var wasConsumed bool
				wasConsumed, capturingPrimitive = handler(action, event, func(p tview.Primitive) {
					a.SetFocus(p)
				})
				if wasConsumed {
					consumed = true
				}
			}
		}
		a.mouseCapturingPrimitive = capturingPrimitive
	}

	x, y := event.Position()
	buttons := event.Buttons()
	clickMoved := x != a.mouseDownX || y != a.mouseDownY
	buttonChanges := buttons ^ a.lastMouseButtons

	if x != a.lastMouseX || y != a.lastMouseY {
		fire(tview.MouseMove)
		a.lastMouseX = x
		a.lastMouseY = y
	}

	for _, buttonEvent := range []struct {
		button                  tcell.ButtonMask
		down, up, click, dclick tview.MouseAction
	}{
		{tcell.ButtonPrimary, tview.MouseLeftDown, tview.MouseLeftUp, tview.MouseLeftClick, tview.MouseLeftDoubleClick},
		{tcell.ButtonMiddle, tview.MouseMiddleDown, tview.MouseMiddleUp, tview.MouseMiddleClick, tview.MouseMiddleDoubleClick},
		{tcell.ButtonSecondary, tview.MouseRightDown, tview.MouseRightUp, tview.MouseRightClick, tview.MouseRightDoubleClick},
	} {
		if buttonChanges&buttonEvent.button != 0 {
			if buttons&buttonEvent.button != 0 {
				fire(buttonEvent.down)
			} else {
				fire(buttonEvent.up) // A user override might set event to nil.
				if !clickMoved && event != nil {
					if a.lastMouseClick.Add(DoubleClickInterval).Before(time.Now()) {
						fire(buttonEvent.click)
						a.lastMouseClick = time.Now()
					} else {
						fire(buttonEvent.dclick)
						a.lastMouseClick = time.Time{} // reset
					}
				}
			}
		}
	}

	for _, wheelEvent := range []struct {
		button tcell.ButtonMask
		action tview.MouseAction
	}{
		{tcell.WheelUp, tview.MouseScrollUp},
		{tcell.WheelDown, tview.MouseScrollDown},
		{tcell.WheelLeft, tview.MouseScrollLeft},
		{tcell.WheelRight, tview.MouseScrollRight}} {
		if buttons&wheelEvent.button != 0 {
			fire(wheelEvent.action)
		}
	}

	return consumed, isMouseDownAction
}
