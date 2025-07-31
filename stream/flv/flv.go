package flv

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/flv"
)

type FLVClient struct {
	url         string
	closer      io.Closer
	demuxer     *flv.Demuxer
	signal      chan any
	packetQueue chan *av.Packet
}

func New(url string) *FLVClient {
	return &FLVClient{
		url:         url,
		signal:      make(chan any),
		packetQueue: make(chan *av.Packet),
	}
}

func (r *FLVClient) Dial() error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Get(r.url)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("status code: %s", resp.Status)
	}

	r.closer = resp.Body
	r.demuxer = flv.NewDemuxer(resp.Body)
	return nil
}

func (r *FLVClient) Close() {
	if r.closer != nil {
		r.closer.Close()
	}
}

func (r *FLVClient) CodecData() ([]av.CodecData, error) {
	streams, err := r.demuxer.Streams()
	if err == nil {
		go func() {
			for {
				packet, err := r.demuxer.ReadPacket()
				if err != nil {
					r.signal <- err
					return
				}
				r.packetQueue <- &packet
			}
		}()
	}
	return streams, err
}

func (r *FLVClient) PacketQueue() <-chan *av.Packet {
	return r.packetQueue
}

func (r *FLVClient) CloseCh() <-chan any {
	return r.signal
}
