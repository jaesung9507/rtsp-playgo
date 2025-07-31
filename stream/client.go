package stream

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"rtsp-playgo/stream/flv"
	"rtsp-playgo/stream/rtmp"
	"rtsp-playgo/stream/rtsp"

	"github.com/deepch/vdk/av"
)

type Client interface {
	Dial() error
	Close()
	CodecData() ([]av.CodecData, error)
	PacketQueue() <-chan *av.Packet
	CloseCh() <-chan any
}

func Dial(streamUrl string) (Client, error) {
	parsedUrl, err := url.Parse(streamUrl)
	if err != nil {
		return nil, err
	}

	var client Client
	switch parsedUrl.Scheme {
	case "rtsp", "rtsps":
		client = rtsp.New(streamUrl)
	case "rtmp", "rtmps":
		client = rtmp.New(streamUrl)
	case "http", "https":
		ext := filepath.Ext(path.Base(parsedUrl.Path))
		switch ext {
		case ".flv":
			client = flv.New(streamUrl)
		default:
			return nil, fmt.Errorf("unsupported extension: %s", ext)
		}
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", parsedUrl.Scheme)
	}

	if err = client.Dial(); err != nil {
		return nil, err
	}

	return client, nil
}
