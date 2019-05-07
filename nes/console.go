package nes

import (
	"encoding/gob"
	"image"
	"image/color"
	"os"
	"path"
)

// Console用于存储NES的相关的信息
type Console struct {
	CPU         *CPU
	APU         *APU
	PPU         *PPU
	Cartridge   *Cartridge
	Controller1 *Controller
	Controller2 *Controller
	Mapper      Mapper
	RAM         []byte
}

// Console作为NES游戏ROM的存储区域
func NewConsole(path string) (*Console, error) {
	// 加载并解析，得到NES游戏文件详细信息
	// PS: 为啥别人解析就能写的那么简单，我就写的那么复杂QAQ
	cartridge, err := LoadNESFile(path)
	if err != nil {
		return nil, err
	}
	ram := make([]byte, 2048)      // 模拟硬件拥有2k内存
	controller1 := NewController() // 控制器1
	controller2 := NewController() // 控制器2
	// 根据游戏文件、控制器和内存属性新建控制台
	console := Console{
		nil, nil, nil, cartridge, controller1, controller2, nil, ram}
	// 根据游戏文件的mapper属性，创建对应的mapper
	mapper, err := NewMapper(&console)
	if err != nil {
		return nil, err
	}
	console.Mapper = mapper
	// center process unit 负责执行程序指令
	console.CPU = NewCPU(&console)
	// audio process unit 负责播放游戏的声音
	console.APU = NewAPU(&console)
	// picture process unit 负责显示游戏的图像
	console.PPU = NewPPU(&console)
	return &console, nil
}

func (console *Console) Reset() {
	console.CPU.Reset()
}

// 游戏运行
func (console *Console) Step() int {
	// CPU运行
	cpuCycles := console.CPU.Step()
	// ???
	ppuCycles := cpuCycles * 3
	for i := 0; i < ppuCycles; i++ {
		// PPU运行
		console.PPU.Step()
		// mapper运行
		console.Mapper.Step()
	}
	for i := 0; i < cpuCycles; i++ {
		// APU运行
		console.APU.Step()
	}
	return cpuCycles
}

// 运行并跳转到PPU的当前帧
func (console *Console) StepFrame() int {
	cpuCycles := 0
	frame := console.PPU.Frame
	for frame == console.PPU.Frame {
		cpuCycles += console.Step()
	}
	return cpuCycles
}

// 根据时间差来决定执行多少个CPU时钟周期
func (console *Console) StepSeconds(seconds float64) {
	cycles := int(CPUFrequency * seconds)
	for cycles > 0 {
		cycles -= console.Step()
	}
}

func (console *Console) Buffer() *image.RGBA {
	return console.PPU.front
}

func (console *Console) BackgroundColor() color.RGBA {
	return Palette[console.PPU.readPalette(0)%64]
}

// 更新按键信息，按键信息是从opengl的pollEvent中获取到的
func (console *Console) SetButtons1(buttons [8]bool) {
	console.Controller1.SetButtons(buttons)
}

// 更新按键信息，按键信息是从opengl的pollEvent中获取到的
func (console *Console) SetButtons2(buttons [8]bool) {
	console.Controller2.SetButtons(buttons)
}

func (console *Console) SetAudioChannel(channel chan float32) {
	console.APU.channel = channel
}

func (console *Console) SetAudioSampleRate(sampleRate float64) {
	if sampleRate != 0 {
		// Convert samples per second to cpu steps per sample
		console.APU.sampleRate = CPUFrequency / sampleRate
		// Initialize filters
		console.APU.filterChain = FilterChain{
			HighPassFilter(float32(sampleRate), 90),
			HighPassFilter(float32(sampleRate), 440),
			LowPassFilter(float32(sampleRate), 14000),
		}
	} else {
		console.APU.filterChain = nil
	}
}

// 保存游戏的当前状态
func (console *Console) SaveState(filename string) error {
	dir, _ := path.Split(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	return console.Save(encoder)
}

// 保存游戏运行所产生的状态信息
func (console *Console) Save(encoder *gob.Encoder) error {
	encoder.Encode(console.RAM)
	console.CPU.Save(encoder)
	console.APU.Save(encoder)
	console.PPU.Save(encoder)
	console.Cartridge.Save(encoder)
	console.Mapper.Save(encoder)
	return encoder.Encode(true)
}

// 从文件中加载游戏状态
func (console *Console) LoadState(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	return console.Load(decoder)
}

// 加载所有游戏运行所需要的状态信息
func (console *Console) Load(decoder *gob.Decoder) error {
	decoder.Decode(&console.RAM)
	console.CPU.Load(decoder)
	console.APU.Load(decoder)
	console.PPU.Load(decoder)
	console.Cartridge.Load(decoder)
	console.Mapper.Load(decoder)
	var dummy bool
	if err := decoder.Decode(&dummy); err != nil {
		return err
	}
	return nil
}
