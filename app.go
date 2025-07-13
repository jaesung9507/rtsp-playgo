package main

import (
	"context"
	"time"

	"github.com/deepch/vdk/format/mp4f"
	"github.com/deepch/vdk/format/rtspv2"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx        context.Context
	rtspClient *rtspv2.RTSPClient
	mp4Muxer   *mp4f.Muxer
	close      chan bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{close: make(chan bool)}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	runtime.EventsOn(a.ctx, "OnUpdateEnd", func(optionalData ...interface{}) {
		a.rtspLoop()
	})
}

func (a *App) MsgBox(msg string) {
	_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Title:   "RTSP PlayGo",
		Message: msg,
		Buttons: []string{"OK"},
	})
}

func (a *App) CloseRTSP() {
	if a.rtspClient != nil {
		a.close <- true
		a.rtspClient.Close()
		a.rtspClient = nil
	}

	if a.mp4Muxer != nil {
		a.mp4Muxer = nil
	}
}

func (a *App) initRTSP(client *rtspv2.RTSPClient, muxer *mp4f.Muxer) {
	a.rtspClient = client
	a.mp4Muxer = muxer
}

func (a *App) rtspLoop() {
	if a.rtspClient != nil && a.mp4Muxer != nil {
		var timeLine = make(map[int8]time.Duration)
		defer runtime.EventsEmit(a.ctx, "OnRTSPStop")
		for {
			select {
			case <-a.close:
				return
			case signals := <-a.rtspClient.Signals:
				switch signals {
				case rtspv2.SignalCodecUpdate:
				case rtspv2.SignalStreamRTPStop:
					return
				}
			case packetAV := <-a.rtspClient.OutgoingPacketQueue:
				timeLine[packetAV.Idx] += packetAV.Duration
				packetAV.Time = timeLine[packetAV.Idx]
				ready, buf, _ := a.mp4Muxer.WritePacket(*packetAV, false)
				if ready {
					runtime.EventsEmit(a.ctx, "OnFrame", buf)
				}
			}
		}
	}
}

func (a *App) RTSP(url string) bool {
	client, err := rtspv2.Dial(rtspv2.RTSPClientOptions{
		URL:              url,
		DisableAudio:     true,
		DialTimeout:      3 * time.Second,
		ReadWriteTimeout: 30 * time.Second,
		Debug:            true,
	})
	if err != nil {
		a.MsgBox(err.Error())
		return false
	}

	muxer := mp4f.NewMuxer(nil)
	if err = muxer.WriteHeader(client.CodecData); err != nil {
		client.Close()
		a.MsgBox(err.Error())
		return false
	}
	meta, init := muxer.GetInit(client.CodecData)
	a.initRTSP(client, muxer)
	runtime.EventsEmit(a.ctx, "OnInit", meta, init)

	return true
}
