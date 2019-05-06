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

// 进入游戏前的操作
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
	// 游戏本身支持持久化的存储数据
	if cartridge.Battery != 0 {
		if sram, err := readSRAM(sramPath(view.hash)); err == nil {
			// 如果可以，则读取save ram的信息
			cartridge.SRAM = sram
		}
	}
}

// 游戏退出前的操作
func (view *GameView) Exit() {
	// 清除回调和声音相关的设置
	view.director.window.SetKeyCallback(nil)
	view.console.SetAudioChannel(nil)
	view.console.SetAudioSampleRate(0)
	// save sram
	cartridge := view.console.Cartridge
	// 游戏本身支持持久化的存储数据
	if cartridge.Battery != 0 {
		// 保存save ram的信息
		writeSRAM(sramPath(view.hash), cartridge.SRAM)
	}
	// save state，退出前保存游戏状态
	view.console.SaveState(savePath(view.hash))
}

// 更新游戏画面
// t  当前的时间戳
// dt 此次执行与上次执行之间的时间差
func (view *GameView) Update(t, dt float64) {
	if dt > 1 {
		dt = 0
	}
	window := view.director.window
	console := view.console
	// 按下摇杆的指定按键，显示menu
	if joystickReset(glfw.Joystick1) {
		view.director.ShowMenu()
	}
	// 按下摇杆的指定按键，显示menu
	if joystickReset(glfw.Joystick2) {
		view.director.ShowMenu()
	}
	// 按下键盘的esc，显示menu
	if readKey(window, glfw.KeyEscape) {
		view.director.ShowMenu()
	}
	// 更新按键信息
	updateControllers(window, console)
	// 根据时间差运行游戏程序~~
	console.StepSeconds(dt)

	// 通过纹理来实现opengl的渲染，纹理本质上就是通过图片进行渲染？
	// 在step的游戏程序运行完毕之后，把texture绑定到gl上
	gl.BindTexture(gl.TEXTURE_2D, view.texture)
	// 根据游戏运行情况，创建一个2d纹理
	setTexture(console.Buffer())
	// 把2d纹理完整的渲染到画布上
	drawBuffer(view.director.window)
	// 清空texture的绑定
	gl.BindTexture(gl.TEXTURE_2D, 0)

	// 如果开启了录像，则保存帧为图片
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

// 通过坐标，把纹理完整的渲染到画布上
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

// 更新按键信息
func updateControllers(window *glfw.Window, console *nes.Console) {
	// 这个turbo是干啥用的？
	turbo := console.PPU.Frame%6 < 3
	k1 := readKeys(window, turbo)
	j1 := readJoystick(glfw.Joystick1, turbo)
	j2 := readJoystick(glfw.Joystick2, turbo)
	console.SetButtons1(combineButtons(k1, j1))
	console.SetButtons2(j2)
}
