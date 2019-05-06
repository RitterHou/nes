package ui

import (
	"log"

	"github.com/fogleman/nes/nes"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

// 显示视图的接口，具体的视图可以实现该接口
// 实际上包含两种视图：1. 菜单视图；2. 游戏视图
type View interface {
	Enter()
	Exit()
	Update(t, dt float64)
}

// 游戏目录
type Director struct {
	window    *glfw.Window // 图形
	audio     *Audio       // 声音
	view      View         // 视图
	menuView  View         // 菜单视图
	timestamp float64
}

func NewDirector(window *glfw.Window, audio *Audio) *Director {
	director := Director{}
	director.window = window
	director.audio = audio
	return &director
}

func (d *Director) SetTitle(title string) {
	d.window.SetTitle(title)
}

// 设置视图
func (d *Director) SetView(view View) {
	// 移除上一个视图
	if d.view != nil {
		d.view.Exit()
	}
	d.view = view
	// 进入新的视图
	if d.view != nil {
		d.view.Enter()
	}
	d.timestamp = glfw.GetTime()
}

// 取指执行，核心引擎
func (d *Director) Step() {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	timestamp := glfw.GetTime()
	dt := timestamp - d.timestamp // 计算此次执行与上次执行之间的时间差
	d.timestamp = timestamp       // 更新执行时间为此次执行的时间戳
	// 这里为啥需要判空？🤔
	if d.view != nil {
		// 更新视图
		d.view.Update(timestamp, dt)
	}
}

func (d *Director) Start(paths []string) {
	// menuView的director为这里的d
	// director的menuView为此menuView
	d.menuView = NewMenuView(d, paths)
	if len(paths) == 1 {
		// 单个游戏则运行游戏
		d.PlayGame(paths[0])
	} else {
		// 多个游戏则显示游戏选择菜单
		d.ShowMenu()
	}
	// 设置好view之后就可以开始执行了
	d.Run()
}

// 开始执行
func (d *Director) Run() {
	// 最核心的死循环
	for !d.window.ShouldClose() {
		d.Step()               // 取指执行
		d.window.SwapBuffers() // 图像切换？
		glfw.PollEvents()      // 接受外界的输入信息，例如按键、手柄，等待
	}
	// 退出时把视图清空
	d.SetView(nil)
}

// 设置游戏的view
func (d *Director) PlayGame(path string) {
	hash, err := hashFile(path)
	if err != nil {
		log.Fatalln(err)
	}
	// 读取游戏并根据游戏的属性设置console的属性
	console, err := nes.NewConsole(path)
	if err != nil {
		log.Fatalln(err)
	}
	// 在游戏界面则设置为游戏view
	d.SetView(NewGameView(d, console, path, hash))
}

// 设置菜单的view
func (d *Director) ShowMenu() {
	d.SetView(d.menuView)
}
