package rtsp

import (
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/rtspv2"
)

type RTSPClient struct {
	url    string
	client *rtspv2.RTSPClient
}

func New(url string) *RTSPClient {
	return &RTSPClient{url: url}
}

func (r *RTSPClient) Dial() error {
	client, err := rtspv2.Dial(rtspv2.RTSPClientOptions{
		URL:                r.url,
		DisableAudio:       false,
		DialTimeout:        3 * time.Second,
		ReadWriteTimeout:   30 * time.Second,
		Debug:              true,
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}
	r.client = client
	return nil
}

func (r *RTSPClient) Close() {
	if r.client != nil {
		r.client.Close()
	}
}

func (r *RTSPClient) CodecData() ([]av.CodecData, error) {
	return r.client.CodecData, nil
}

func (r *RTSPClient) PacketQueue() <-chan *av.Packet {
	return r.client.OutgoingPacketQueue
}

func (r *RTSPClient) CloseCh() <-chan interface{} {
	closeCh := make(chan interface{})
	go func() {
		for s := range r.client.Signals {
			if s == rtspv2.SignalStreamRTPStop {
				closeCh <- s
				return
			}
		}
	}()
	return closeCh
}
