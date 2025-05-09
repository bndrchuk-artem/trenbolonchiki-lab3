package ui

import (
	"image"
	"image/color"
	"log"
	"sync"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/draw"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

var DefaultWindowSize = image.Pt(800, 800)

type Visualizer struct {
	Title         string
	Debug         bool
	OnScreenReady func(s screen.Screen)

	w    screen.Window
	tx   chan screen.Texture
	done chan struct{}

	sz       size.Event
	pos      image.Rectangle
	mousePos image.Point

	mu sync.Mutex
}

// Main starts the main application window and event loop
func (pw *Visualizer) Main() {
	pw.tx = make(chan screen.Texture, 1) // Buffer to prevent blocking
	pw.done = make(chan struct{})
	pw.pos.Max.X = 200
	pw.pos.Max.Y = 200
	driver.Main(pw.run)
}

// Update receives a new texture to display
func (pw *Visualizer) Update(t screen.Texture) {
	select {
	case pw.tx <- t:
	default:
		if pw.w != nil {
			pw.w.Send(paint.Event{})
		}
		pw.tx <- t
	}
}

func (pw *Visualizer) run(s screen.Screen) {
	w, err := s.NewWindow(&screen.NewWindowOptions{
		Title:  pw.Title,
		Width:  DefaultWindowSize.X,
		Height: DefaultWindowSize.Y,
	})
	if err != nil {
		log.Fatal("Failed to initialize the app window:", err)
	}
	defer func() {
		w.Release()
		close(pw.done)
	}()

	pw.w = w

	if pw.OnScreenReady != nil {
		pw.OnScreenReady(s)
	}

	events := make(chan any, 100)
	go func() {
		for {
			e := w.NextEvent()
			if pw.Debug {
				log.Printf("Event: %T %v", e, e)
			}

			if detectTerminate(e) {
				close(events)
				return
			}
			events <- e
		}
	}()

	var t screen.Texture

	for {
		select {
		case e, ok := <-events:
			if !ok {
				return
			}
			pw.handleEvent(e, t)

		case t = <-pw.tx:
			if pw.w != nil {
				pw.w.Send(paint.Event{})
			}
		}
	}
}

func detectTerminate(e any) bool {
	switch e := e.(type) {
	case lifecycle.Event:
		if e.To == lifecycle.StageDead {
			return true
		}
	case key.Event:
		if e.Code == key.CodeEscape || (e.Code == key.CodeQ && e.Modifiers == key.ModControl) {
			return true
		}
	}
	return false
}

func (pw *Visualizer) handleEvent(e any, t screen.Texture) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	switch e := e.(type) {
	case size.Event:
		pw.sz = e

		if pw.w != nil {
			pw.w.Send(paint.Event{})
		}

	case error:
		log.Printf("ERROR: %s", e)

	case mouse.Event:
		switch e.Direction {
		case mouse.DirPress:
			if e.Button == mouse.ButtonRight {
				pw.mousePos = image.Point{X: int(e.X), Y: int(e.Y)}
				if pw.w != nil {
					pw.w.Send(paint.Event{})
				}
			}
		}

	case paint.Event:
		if pw.w == nil {
			return
		}

		if t == nil {
			pw.drawDefaultUI()
		} else {
			pw.w.Scale(pw.sz.Bounds(), t, t.Bounds(), draw.Src, nil)
		}
		pw.w.Publish()
	}
}

// drawDefaultUI draws the default UI when no texture is available
func (pw *Visualizer) drawDefaultUI() {
	if pw.w == nil {
		return
	}

	// Fill background
	pw.w.Fill(pw.sz.Bounds(), color.White, draw.Src)

	// Determine center position
	var centerX, centerY int
	if pw.mousePos == (image.Point{}) {
		centerX, centerY = pw.sz.WidthPx/2, pw.sz.HeightPx/2
	} else {
		centerX, centerY = pw.mousePos.X, pw.mousePos.Y
	}

	// Size parameters
	horizontalBarWidth := 400
	horizontalBarHeight := 100
	verticalBarWidth := horizontalBarWidth / 4

	// Color
	blue := color.RGBA{R: 0, G: 0, B: 255, A: 255}

	// Draw T-shape figure
	horizontalBar := image.Rect(
		centerX-horizontalBarWidth/2,
		centerY,
		centerX+horizontalBarWidth/2,
		centerY+horizontalBarHeight,
	)

	verticalBar := image.Rect(
		centerX-verticalBarWidth/2,
		centerY-200,
		centerX+verticalBarWidth/2,
		centerY,
	)

	// Draw the parts
	pw.w.Fill(verticalBar, blue, draw.Src)
	pw.w.Fill(horizontalBar, blue, draw.Src)
}
