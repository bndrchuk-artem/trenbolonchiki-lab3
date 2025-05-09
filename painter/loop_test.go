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

func TestLoop_Post_Success(t *testing.T) {
	textureMock := new(Mock)
	receiverMock := new(Mock)
	screenMock := new(Mock)

	texture := image.Pt(800, 800)
	screenMock.On("NewTexture", texture).Return(textureMock, nil)
	receiverMock.On("Update", textureMock).Return()
	loop := Loop{
		Receiver: receiverMock,
	}

	loop.Start(screenMock)

	operationOne := new(Mock)
	textureMock.On("Bounds").Return(image.Rectangle{})
	operationOne.On("Do", textureMock).Return(true)

	assert.Empty(t, loop.MsgQueue.Queue)
	loop.Post(operationOne)
	time.Sleep(1 * time.Second)
	assert.Empty(t, loop.MsgQueue.Queue)

	operationOne.AssertCalled(t, "Do", textureMock)
	receiverMock.AssertCalled(t, "Update", textureMock)
	screenMock.AssertCalled(t, "NewTexture", image.Pt(800, 800))
}

func TestLoop_Post_Failure(t *testing.T) {
	textureMock := new(Mock)
	receiverMock := new(Mock)
	screenMock := new(Mock)

	texture := image.Pt(800, 800)
	screenMock.On("NewTexture", texture).Return(textureMock, nil)
	receiverMock.On("Update", textureMock).Return()
	loop := Loop{
		Receiver: receiverMock,
	}

	loop.Start(screenMock)

	operationOne := new(Mock)
	textureMock.On("Bounds").Return(image.Rectangle{})
	operationOne.On("Do", textureMock).Return(false)

	assert.Empty(t, loop.MsgQueue.Queue)
	loop.Post(operationOne)
	time.Sleep(1 * time.Second)
	assert.Empty(t, loop.MsgQueue.Queue)

	operationOne.AssertCalled(t, "Do", textureMock)
	receiverMock.AssertNotCalled(t, "Update", textureMock)
	screenMock.AssertCalled(t, "NewTexture", image.Pt(800, 800))
}

func TestLoop_Post_Multiple_Success(t *testing.T) {
	textureMock := new(Mock)
	receiverMock := new(Mock)
	screenMock := new(Mock)

	texture := image.Pt(800, 800)
	screenMock.On("NewTexture", texture).Return(textureMock, nil)
	receiverMock.On("Update", textureMock).Return()
	loop := Loop{
		Receiver: receiverMock,
	}

	loop.Start(screenMock)

	operationOne := new(Mock)
	operationTwo := new(Mock)
	textureMock.On("Bounds").Return(image.Rectangle{})
	operationOne.On("Do", textureMock).Return(true)
	operationTwo.On("Do", textureMock).Return(true)

	assert.Empty(t, loop.MsgQueue.Queue)
	loop.Post(operationOne)
	loop.Post(operationTwo)
	time.Sleep(1 * time.Second)
	assert.Empty(t, loop.MsgQueue.Queue)

	operationOne.AssertCalled(t, "Do", textureMock)
	operationTwo.AssertCalled(t, "Do", textureMock)
	receiverMock.AssertCalled(t, "Update", textureMock)
	screenMock.AssertCalled(t, "NewTexture", image.Pt(800, 800))
}

