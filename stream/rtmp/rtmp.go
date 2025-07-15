package rtmp

import (
	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/rtmp"
)

type RTMPClient struct {
	url         string
	conn        *rtmp.Conn
	signal      chan interface{}
	packetQueue chan *av.Packet
}

func New(url string) *RTMPClient {
	return &RTMPClient{
		url:         url,
		signal:      make(chan interface{}),
		packetQueue: make(chan *av.Packet),
	}
}

func (r *RTMPClient) Dial() error {
	conn, err := rtmp.Dial(r.url)
	if err != nil {
		return err
	}
	r.conn = conn
	return nil
}

func (r *RTMPClient) Close() {
	if r.conn != nil {
		r.conn.Close()
	}
}

func (r *RTMPClient) CodecData() ([]av.CodecData, error) {
	streams, err := r.conn.Streams()
	if err == nil {
		go func() {
			for {
				packet, err := r.conn.ReadPacket()
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

func (r *RTMPClient) PacketQueue() <-chan *av.Packet {
	return r.packetQueue
}

func (r *RTMPClient) CloseCh() <-chan interface{} {
	return r.signal
}
