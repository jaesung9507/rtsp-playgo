package http

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/flv"
	"github.com/deepch/vdk/format/ts"
)

type HTTPClient struct {
	url         *url.URL
	closer      io.Closer
	demuxer     av.Demuxer
	signal      chan any
	packetQueue chan *av.Packet
}

func New(parsedUrl *url.URL) *HTTPClient {
	return &HTTPClient{
		url:         parsedUrl,
		signal:      make(chan any),
		packetQueue: make(chan *av.Packet),
	}
}

func (h *HTTPClient) getDemuxerFunc() (func(r io.Reader) av.Demuxer, error) {
	ext := filepath.Ext(path.Base(h.url.Path))
	switch ext {
	case ".flv":
		return func(r io.Reader) av.Demuxer { return flv.NewDemuxer(r) }, nil
	case ".ts":
		return func(r io.Reader) av.Demuxer { return ts.NewDemuxer(r) }, nil
	}
	return nil, fmt.Errorf("unsupported extension: %s", ext)
}

func (h *HTTPClient) Dial() error {
	newDemuxer, err := h.getDemuxerFunc()
	if err != nil {
		return err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Get(h.url.String())
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("status code: %s", resp.Status)
	}

	h.closer = resp.Body
	h.demuxer = newDemuxer(resp.Body)
	return nil
}

func (h *HTTPClient) Close() {
	if h.closer != nil {
		h.closer.Close()
	}
}

func (h *HTTPClient) CodecData() ([]av.CodecData, error) {
	streams, err := h.demuxer.Streams()
	if err == nil {
		go func() {
			for {
				packet, err := h.demuxer.ReadPacket()
				if err != nil {
					h.signal <- err
					return
				}
				h.packetQueue <- &packet
			}
		}()
	}
	return streams, err
}

func (h *HTTPClient) PacketQueue() <-chan *av.Packet {
	return h.packetQueue
}

func (h *HTTPClient) CloseCh() <-chan any {
	return h.signal
}
