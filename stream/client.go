package stream

import (
	"fmt"
	"net/url"

	"github.com/jaesung9507/playgo/stream/http"
	"github.com/jaesung9507/playgo/stream/rtmp"
	"github.com/jaesung9507/playgo/stream/rtsp"

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
		client = rtsp.New(parsedUrl)
	case "rtmp", "rtmps":
		client = rtmp.New(parsedUrl)
	case "http", "https":
		client = http.New(parsedUrl)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", parsedUrl.Scheme)
	}

	if err = client.Dial(); err != nil {
		return nil, err
	}

	return client, nil
}
