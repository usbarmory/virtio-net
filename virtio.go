// VirtIO network driver
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package vnet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/usbarmory/tamago/kvm/virtio"
)

// Device parameters
const (
	DeviceID   = 0x01
	ConfigSize = 24
)

// virtual queue pairs
const (
	// receiveq1
	rxq = 0
	// transmitq1
	txq = 1
)

// Configuration constants
const (
	STATUS_DOWN = 0
	STATUS_UP   = 1

	SPEED_MIN = 0x00000000
	SPEED_MAX = 0x7fffffff

	DUPLEX_HALF = 0x00
	DUPLEX_FULL = 0x01

	HASH_TYPE_IPV4   = 1 << 0
	HASH_TYPE_TCPV4  = 1 << 1
	HASH_TYPE_UDPV4  = 1 << 2
	HASH_TYPE_IPV6   = 1 << 3
	HASH_TYPE_TCPV6  = 1 << 4
	HASH_TYPE_UDPV6  = 1 << 5
	HASH_TYPE_IP_EX  = 1 << 6
	HASH_TYPE_TCP_EX = 1 << 7
	HASH_TYPE_UDP_EX = 1 << 8
)

// Supported Features
const (
	FeatureChecksum    = (1 << 0)
	FeatureMTU         = (1 << 3)
	FeatureMAC         = (1 << 5)
	FeatureStatus      = (1 << 16)
	FeatureSpeedDuplex = (1 << 63)

	DriverFeatures = FeatureChecksum | FeatureMTU | FeatureMAC | FeatureStatus | FeatureSpeedDuplex
)

// Header flags
const (
	NeedsChecksum = 0
)

// Header represents a VirtIO network device header (virtio_net_hdr)
type Header struct {
	Flags      uint8
	GSOType    uint8
	HdrLen     uint16
	GSOSize    uint16
	CSumStart  uint16
	CSumOffset uint16
	NumBuffers uint16 // not used in legacy drivers
}

// HeaderLength defines the VirtIO network device header length, given that
// certain legacy implementations (e.g. QEMU) omit certain fields, and that the
// header is not used by the driver, its length is defined as a variable.
var HeaderLength = 10

// Bytes converts the descriptor structure to byte array format.
func (d *Header) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes()
}

// Config represents a VirtIO network device configuration
type Config struct {
	// MAC represents the interface physical address.
	MAC [6]byte
	// Status represents the driver status.
	Status uint16
	// MaxVirtualQueuePairs is the Maximum number of tx/rx queues.
	MaxVirtualQueuePairs uint16
	// MTU represents the Ethernet Maximum Transmission Unit.
	MTU uint16
	// Speed represents the device speed in units of 1Mbps.
	Speed uint32
	// Duplex represents the communication mode.
	Duplex uint8
	// RSSMaxKeySize represents the Receive Side Scaling hash maximum key size.
	RSSMaxKeySize uint8
	// RSSMaxIndirectionTableLength represents the Receive Side Scaling maximum indirection table length.
	RSSMaxIndirectionTableLength uint16
	// SupportedHashTypes represents the supported hash types.
	SupportedHashTypes uint32
}

// Net represents a VirtIO network device instance.
type Net struct {
	sync.Mutex

	// Controller index
	Index int
	// Base register
	Base uint32
	// Interrupt ID
	IRQ int

	// Incoming packet handler
	RxHandler func([]byte)
	// Maximum Transmission Unit
	MTU uint16

	// VirtIO instance
	io *virtio.VirtIO

	// receive queue
	rx *virtio.VirtualQueue
	// transmit queue
	tx *virtio.VirtualQueue
}

func (hw *Net) initQueue(index int, flags uint16) (queue *virtio.VirtualQueue) {
	size := hw.io.MaxQueueSize(index)
	length := hw.MTU + uint16(HeaderLength)

	queue = &virtio.VirtualQueue{}
	queue.Init(size, int(length), flags)

	hw.io.SetQueueSize(index, size)

	return
}

// Init initializes the VirtIO network device.
func (hw *Net) Init() (err error) {
	hw.Lock()
	defer hw.Unlock()

	hw.io = &virtio.VirtIO{
		Base:       hw.Base,
		ConfigSize: ConfigSize,
	}

	if err := hw.io.Init(DriverFeatures); err != nil {
		return err
	}

	if id := hw.io.DeviceID(); id != DeviceID {
		return fmt.Errorf("incompatible device ID (%x != DeviceID)", id, DeviceID)
	}

	if mtu := hw.Config().MTU; hw.MTU > mtu {
		return fmt.Errorf("incompatible MTU ID (%d > %d)", hw.MTU, mtu)
	}

	if hw.io.QueueReady(rxq) || hw.io.QueueReady(txq) {
		return errors.New("queues unavailable")
	}

	hw.rx = hw.initQueue(rxq, virtio.Write)
	hw.tx = hw.initQueue(txq, 0)

	return
}

// Config returns the network device configuration.
func (hw *Net) Config() (config Config) {
	if hw.io == nil || len(hw.io.Config) != ConfigSize {
		return
	}

	binary.Decode(hw.io.Config, binary.LittleEndian, &config)

	return
}

// Start begins processing of incoming packets. When the argument is true the
// function waits and handles received packets (see Rx()) through RxHandler()
// (when set), it should never return.
func (hw *Net) Start(rx bool) {
	var buf []byte

	if hw.rx == nil || hw.tx == nil {
		return
	}

	hw.io.SetQueue(rxq, hw.rx)
	hw.io.SetQueue(txq, hw.tx)

	hw.io.QueueNotify(rxq)

	if !rx || hw.RxHandler == nil {
		return
	}

	for {
		runtime.Gosched()

		if buf = hw.Rx(); buf != nil {
			hw.RxHandler(buf)
		}
	}
}

// Rx receives a single network frame, excluding the checksum, from the MAC
// controller ring buffer.
func (hw *Net) Rx() []byte {
	buf := hw.rx.Pop()

	if len(buf) < HeaderLength {
		return nil
	}

	return buf[HeaderLength:]
}

// Tx transmits a single network frame, the checksum is appended automatically
// and must not be included.
func (hw *Net) Tx(buf []byte) {
	hdr := make([]byte, HeaderLength)
	buf = append(hdr, buf...)

	hw.tx.Push(buf)
	hw.io.QueueNotify(txq)
}
