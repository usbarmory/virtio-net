// VirtIO network driver
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package vnet

import (
	"encoding/binary"
	"fmt"
	"runtime"
	"sync"

	"github.com/usbarmory/tamago/virtio"
)

// VirtIO configuration
const (
	DeviceID   = 0x01
	ConfigSize = 24
)

// Config constants
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

	// Incoming packet handler
	RxHandler func([]byte)

	// VirtIO instance
	io *virtio.VirtIO
}

// Init initializes the VirtIO network device.
func (hw *Net) Init() (err error) {
	hw.Lock()
	defer hw.Unlock()

	hw.io = &virtio.VirtIO{
		Base:       hw.Base,
		ConfigSize: ConfigSize,
	}

	if err := hw.io.Init(); err != nil {
		return err
	}

	if id := hw.io.DeviceID(); id != DeviceID {
		return fmt.Errorf("incompatible device ID (%x)", id)
	}

	// receiveq1
	hw.io.SelectQueue(0)
	hw.io.SetQueueSize(hw.io.MaxQueueSize())

	// transmitq1
	hw.io.SelectQueue(1)
	hw.io.SetQueueSize(hw.io.MaxQueueSize())

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
func (hw *Net) Rx() (buf []byte) {
	hw.Lock()
	defer hw.Unlock()

	// TODO

	return
}

// Tx transmits a single network frame, the checksum is appended automatically
// and must not be included.
func (hw *Net) Tx(buf []byte) {
	hw.Lock()
	defer hw.Unlock()

	// TODO
}
