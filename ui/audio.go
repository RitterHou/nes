package ui

import "github.com/gordonklaus/portaudio"

// 声音相关的配置
type Audio struct {
	stream         *portaudio.Stream // 跨平台的声音库
	sampleRate     float64
	outputChannels int
	channel        chan float32
}

func NewAudio() *Audio {
	a := Audio{}
	a.channel = make(chan float32, 44100)
	return &a
}

func (a *Audio) Start() error {
	// 获取到默认的声音设备
	host, err := portaudio.DefaultHostApi()
	if err != nil {
		return err
	}
	parameters := portaudio.HighLatencyParameters(nil, host.DefaultOutputDevice)
	stream, err := portaudio.OpenStream(parameters, a.Callback)
	if err != nil {
		return err
	}
	if err := stream.Start(); err != nil {
		return err
	}
	a.stream = stream
	a.sampleRate = parameters.SampleRate
	a.outputChannels = parameters.Output.Channels
	return nil
}

func (a *Audio) Stop() error {
	return a.stream.Close()
}

func (a *Audio) Callback(out []float32) {
	var output float32
	for i := range out {
		if i%a.outputChannels == 0 {
			select {
			case sample := <-a.channel:
				output = sample
			default:
				output = 0
			}
		}
		out[i] = output
	}
}
