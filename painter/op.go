package painter

import (
	"image"
	"image/color"

	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/draw"
)

type Operation interface {
	Do(t screen.Texture) (ready bool)
}

type OperationList []Operation

func (ol OperationList) Do(t screen.Texture) (ready bool) {
	for _, o := range ol {
		if o == nil {
			continue
		}
		ready = o.Do(t) || ready
	}
	return
}

// UpdateOp операція, яка не змінює текстуру, але сигналізує, що текстуру потрібно розглядати як готову
var UpdateOp = updateOp{}

type updateOp struct{}

func (op updateOp) Do(t screen.Texture) bool { return true }

// OperationFunc використовується для перетворення функції оновлення текстури в Operation
type OperationFunc func(t screen.Texture)

func (f OperationFunc) Do(t screen.Texture) bool {
	f(t)
	return false
}

func WhiteFill(t screen.Texture) {
	t.Fill(t.Bounds(), color.White, screen.Src)
}

func GreenFill(t screen.Texture) {
	t.Fill(t.Bounds(), color.RGBA{G: 0xff, A: 0xff}, screen.Src)
}

type BgRectangle struct {
	X1, Y1, X2, Y2 int
}

func (op *BgRectangle) Do(t screen.Texture) bool {
	if op == nil || t == nil {
		return false
	}

	x1, y1 := op.X1, op.Y1
	x2, y2 := op.X2, op.Y2

	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	bounds := t.Bounds()
	if x1 < bounds.Min.X {
		x1 = bounds.Min.X
	}
	if y1 < bounds.Min.Y {
		y1 = bounds.Min.Y
	}
	if x2 > bounds.Max.X {
		x2 = bounds.Max.X
	}
	if y2 > bounds.Max.Y {
		y2 = bounds.Max.Y
	}

	blue := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	t.Fill(image.Rect(x1, y1, x2, y2), blue, draw.Src)
	return false
}

type Figure struct {
	X, Y int
	C    color.RGBA
}

func (op *Figure) Do(t screen.Texture) bool {
	if op == nil || t == nil {
		return false
	}

	figureColor := op.C
	if figureColor.A == 0 {
		figureColor = color.RGBA{R: 0, G: 0, B: 255, A: 255}
	}

	bounds := t.Bounds()

	horizontalX1 := op.X - 200
	horizontalY1 := op.Y
	horizontalX2 := op.X + 200
	horizontalY2 := op.Y + 100

	verticalX1 := op.X - 62
	verticalY1 := op.Y - 200
	verticalX2 := op.X + 63
	verticalY2 := op.Y

	if isRectVisible(horizontalX1, horizontalY1, horizontalX2, horizontalY2, bounds) {
		t.Fill(image.Rect(horizontalX1, horizontalY1, horizontalX2, horizontalY2), figureColor, draw.Src)
	}

	if isRectVisible(verticalX1, verticalY1, verticalX2, verticalY2, bounds) {
		t.Fill(image.Rect(verticalX1, verticalY1, verticalX2, verticalY2), figureColor, draw.Src)
	}

	return false
}

func isRectVisible(x1, y1, x2, y2 int, bounds image.Rectangle) bool {
	return x1 < bounds.Max.X && x2 > bounds.Min.X && y1 < bounds.Max.Y && y2 > bounds.Min.Y
}

type Move struct {
	X, Y    int
	Figures []*Figure
}

func (op *Move) Do(t screen.Texture) bool {
	if op == nil || op.Figures == nil {
		return false
	}

	for i := range op.Figures {
		if op.Figures[i] != nil {
			op.Figures[i].X += op.X
			op.Figures[i].Y += op.Y
		}
	}
	return false
}

func ResetScreen(t screen.Texture) {
	if t != nil {
		t.Fill(t.Bounds(), color.RGBA{0, 0, 0, 255}, draw.Src)
	}
}
