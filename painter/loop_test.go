package painter


import (
	"image"
	"image/color"
	"image/draw"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/exp/shiny/screen"
)

type Mock struct {
	mock.Mock
}

func (_ *Mock) NewBuffer(size image.Point) (screen.Buffer, error) {
	return nil, nil
}

func (_ *Mock) NewWindow(opts *screen.NewWindowOptions) (screen.Window, error) {
	return nil, nil
}

func (mockReceiver *Mock) Update(texture screen.Texture) {
	mockReceiver.Called(texture)
}

func (mockScreen *Mock) NewTexture(size image.Point) (screen.Texture, error) {
	args := mockScreen.Called(size)
	return args.Get(0).(screen.Texture), args.Error(1)
}

func (mockTexture *Mock) Release() {
	mockTexture.Called()
}

func (mockTexture *Mock) Upload(dp image.Point, src screen.Buffer, sr image.Rectangle) {
	mockTexture.Called(dp, src, sr)
}

func (mockTexture *Mock) Bounds() image.Rectangle {
	args := mockTexture.Called()
	return args.Get(0).(image.Rectangle)
}

func (mockTexture *Mock) Fill(dr image.Rectangle, src color.Color, op draw.Op) {
	mockTexture.Called(dr, src, op)
}

func (mockTexture *Mock) Size() image.Point {
	args := mockTexture.Called()
	return args.Get(0).(image.Point)
}

func (mockOperation *Mock) Do(t screen.Texture) bool {
	args := mockOperation.Called(t)
	return args.Bool(0)
}

func TestLoop_Post(t *testing.T) {
	var (
		l  Loop
		tr testReceiver
	)
	l.Receiver = &tr

	var testOps []string

	l.Start(mockScreen{})
	l.Post(logOp(t, "do white fill", WhiteFill))
	l.Post(logOp(t, "do green fill", GreenFill))
	l.Post(UpdateOp)

	for i := 0; i < 3; i++ {
		go l.Post(logOp(t, "do green fill", GreenFill))
	}

	l.Post(OperationFunc(func(screen.Texture) {
		testOps = append(testOps, "op 1")
		l.Post(OperationFunc(func(screen.Texture) {
			testOps = append(testOps, "op 2")
		}))
	}))
	l.Post(OperationFunc(func(screen.Texture) {
		testOps = append(testOps, "op 3")
	}))

	l.StopAndWait()

	if tr.lastTexture == nil {
		t.Fatal("Texture was not updated")
	}
	mt, ok := tr.lastTexture.(*mockTexture)
	if !ok {
		t.Fatal("Unexpected texture", tr.lastTexture)
	}
	if mt.Colors[0] != color.White {
		t.Error("First color is not white:", mt.Colors)
	}
	if len(mt.Colors) != 2 {
		t.Error("Unexpected size of colors:", mt.Colors)
	}

	if !reflect.DeepEqual(testOps, []string{"op 1", "op 2", "op 3"}) {
		t.Error("Bad order:", testOps)
	}
}

func logOp(t *testing.T, msg string, op OperationFunc) OperationFunc {
	return func(tx screen.Texture) {
		t.Log(msg)
		op(tx)
	}
}


