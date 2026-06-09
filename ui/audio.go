package ui

import (
	"errors"
	"unsafe"

	"github.com/Zyko0/go-sdl3/sdl"
)

const (
	AudioSampleRate = 44100
	AudioBufferSize = 256
	AudioSampleSize = 2
)

type Audio struct {
	stream  *sdl.AudioStream
	channel chan float32
	volume  float32
}

func NewAudio(volume int) *Audio {
	return &Audio{
		channel: make(chan float32, AudioSampleRate),
		volume:  min(max(0, float32(volume)/255), 1),
	}
}

func (a *Audio) Start() error {
	spec := &sdl.AudioSpec{
		Format:   sdl.AUDIO_S16,
		Channels: 1,
		Freq:     AudioSampleRate,
	}
	callback := sdl.NewAudioStreamCallback(a.callback)
	a.stream = sdl.AUDIO_DEVICE_DEFAULT_PLAYBACK.OpenAudioDeviceStream(spec, callback)
	if a.stream == nil {
		return errors.New("error opening default audio stream")
	}
	return a.stream.ResumeDevice()
}

func (a *Audio) Stop() error {
	a.stream.Destroy()
	return nil
}

func (a *Audio) callback(stream *sdl.AudioStream, neededBytes, totalBytes int32) {
	var buffer [AudioBufferSize]int16
	needed := int(neededBytes / AudioSampleSize)
	for needed > 0 {
		n := min(needed, AudioBufferSize)
		for i := range n {
			select {
			case sample := <-a.channel:
				buffer[i] = int16(32767 * sample * a.volume)
			default:
				buffer[i] = 0
			}
		}
		stream.PutData(asBytes(buffer[:n]))
		needed -= n
	}
}

func asBytes(s []int16) []uint8 {
	return unsafe.Slice((*uint8)(unsafe.Pointer(&s[0])), len(s)*2)
}
