// VirtIO network driver
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package vnet

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"runtime"
	"sync"

	"github.com/usbarmory/tamago/virtio"
)

// VirtIO device type
const NetworkCard = 0x01

// Net represents a VirtIO network device instance.
type Net struct {
	sync.Mutex

	// Controller index
	Index int
	// Base register
	Base uint32

	// MAC address (use SetMAC() for post Init() changes)
	MAC net.HardwareAddr
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
		Base: hw.Base,
	}

	if hw.MAC == nil {
		hw.MAC = make([]byte, 6)
		rand.Read(hw.MAC)
		// flag address as unicast and locally administered
		hw.MAC[0] &= 0xfe
		hw.MAC[0] |= 0x02
	} else if len(hw.MAC) != 6 {
		return errors.New("invalid MAC")
	}

	if err := hw.io.Init(); err != nil {
		return err
	}

	if id := hw.io.DeviceID(); id != NetworkCard {
		return fmt.Errorf("incompatible device ID (%x)", id)
	}

	hw.io.SelectQueue(0)

	// set physical address
	hw.SetMAC(hw.MAC)

	return
}

// SetMAC allows to change the controller physical address register after
// initialization.
func (hw *Net) SetMAC(mac net.HardwareAddr) {
	hw.MAC = mac
	// FIXME: TODO
}

// Start begins processing of incoming packets. When the argument is true the
// function waits and handles received packets (see Rx()) through RxHandler()
// (when set), it should never return.
func (hw *Net) Start(rx bool) {
	if !rx || hw.RxHandler == nil {
		return
	}

	var buf []byte

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

	// FIXME: TODO

	return
}

// Tx transmits a single network frame, the checksum is appended automatically
// and must not be included.
func (hw *Net) Tx(buf []byte) {
	hw.Lock()
	defer hw.Unlock()

	// FIXME: TODO
}
