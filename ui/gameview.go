package ui

import (
	"image"

	"github.com/fogleman/nes/nes"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

const padding = 0

type GameView struct {
	director *Director
	console  *nes.Console
	title    string
	hash     string
	texture  uint32
	record   bool
	frames   []image.Image
}

// 创建游戏view
func NewGameView(director *Director, console *nes.Console, title, hash string) View {
	texture := createTexture()
	return &GameView{director, console, title, hash, texture, false, nil}
}

func (view *GameView) Enter() {
	gl.ClearColor(0, 0, 0, 1)
	view.director.SetTitle(view.title) // 视图的title其实就是游戏的path
	// 添加声音的相关设置
	view.console.SetAudioChannel(view.director.audio.channel)
	view.console.SetAudioSampleRate(view.director.audio.sampleRate)
	// 设置按键回调
	view.director.window.SetKeyCallback(view.onKey)
	// load state，读取已经保存的游戏状态
	if err := view.console.LoadState(savePath(view.hash)); err == nil {
		// 如果读取成功则直接返回
		return
	} else {
		// 未能读取到游戏状态数据则需要进行初始化
		view.console.Reset()
	}
	// load sram
	cartridge := view.console.Cartridge
	if cartridge.Battery != 0 {
		if sram, err := readSRAM(sramPath(view.hash)); err == nil {
			cartridge.SRAM = sram
		}
	}
}

func (view *GameView) Exit() {
	view.director.window.SetKeyCallback(nil)
	view.console.SetAudioChannel(nil)
	view.console.SetAudioSampleRate(0)
	// save sram
	cartridge := view.console.Cartridge
	if cartridge.Battery != 0 {
		writeSRAM(sramPath(view.hash), cartridge.SRAM)
	}
	// save state，退出前保存游戏状态
	view.console.SaveState(savePath(view.hash))
}

// 更新游戏画面
func (view *GameView) Update(t, dt float64) {
	if dt > 1 {
		dt = 0
	}
	window := view.director.window
	console := view.console
	if joystickReset(glfw.Joystick1) {
		view.director.ShowMenu()
	}
	if joystickReset(glfw.Joystick2) {
		view.director.ShowMenu()
	}
	if readKey(window, glfw.KeyEscape) {
		view.director.ShowMenu()
	}
	updateControllers(window, console)
	// 设置时间差
	console.StepSeconds(dt)
	gl.BindTexture(gl.TEXTURE_2D, view.texture)
	// 设置图像内容
	setTexture(console.Buffer())
	// 在屏幕上面画出图像
	drawBuffer(view.director.window)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	if view.record {
		view.frames = append(view.frames, copyImage(console.Buffer()))
	}
}

// 游戏视图的按键回调
func (view *GameView) onKey(window *glfw.Window,
	key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		switch key {
		case glfw.KeySpace:
			// 截屏
			screenshot(view.console.Buffer())
		case glfw.KeyR:
			// 重新开始游戏
			view.console.Reset()
		case glfw.KeyTab:
			// 录像并保存为gif图片
			if view.record {
				view.record = false
				animation(view.frames)
				view.frames = nil
			} else {
				view.record = true
			}
		}
	}
}

func drawBuffer(window *glfw.Window) {
	w, h := window.GetFramebufferSize()
	s1 := float32(w) / 256
	s2 := float32(h) / 240
	f := float32(1 - padding)
	var x, y float32
	if s1 >= s2 {
		x = f * s2 / s1
		y = f
	} else {
		x = f
		y = f * s1 / s2
	}
	gl.Begin(gl.QUADS)
	gl.TexCoord2f(0, 1)
	gl.Vertex2f(-x, -y)
	gl.TexCoord2f(1, 1)
	gl.Vertex2f(x, -y)
	gl.TexCoord2f(1, 0)
	gl.Vertex2f(x, y)
	gl.TexCoord2f(0, 0)
	gl.Vertex2f(-x, y)
	gl.End()
}

func updateControllers(window *glfw.Window, console *nes.Console) {
	turbo := console.PPU.Frame%6 < 3
	k1 := readKeys(window, turbo)
	j1 := readJoystick(glfw.Joystick1, turbo)
	j2 := readJoystick(glfw.Joystick2, turbo)
	console.SetButtons1(combineButtons(k1, j1))
	console.SetButtons2(j2)
}
