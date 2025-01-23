// VirtIO network driver
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package vnet

import (
	"encoding/binary"
	"errors"
	"net"

	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// NIC represents an virtual Ethernet instance.
type NIC struct {
	// Link is a gVisor channel endpoint
	Link *channel.Endpoint

	// Device is the physical interface associated to the virtual one.
	Device *Net

	// MAC address
	mac net.HardwareAddr
}

type notification struct {
	nic *NIC
}

func (n *notification) WriteNotify() {
	n.nic.Device.Tx(n.nic.Tx())
}

// Init initializes a virtual Ethernet instance bound to a physical Ethernet
// device.
func (nic *NIC) Init() (err error) {
	if nic.Link == nil {
		return errors.New("missing link endpoint")
	}

	if nic.Device == nil {
		return
	}

	nic.Device.RxHandler = nic.Rx

	nic.Link.AddNotify(&notification{
		nic: nic,
	})

	return
}

// Rx receives a single Ethernet frame from the virtual Ethernet instance.
func (nic *NIC) Rx(buf []byte) {
	if len(buf) < 14 {
		return
	}

	hdr := buf[0:14]
	proto := tcpip.NetworkProtocolNumber(binary.BigEndian.Uint16(buf[12:14]))
	payload := buf[14:]

	pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
		ReserveHeaderBytes: len(hdr),
		Payload:            buffer.MakeWithData(payload),
	})

	copy(pkt.LinkHeader().Push(len(hdr)), hdr)

	nic.Link.InjectInbound(proto, pkt)

	return
}

// Tx transmits a single Ethernet frame to the virtual Ethernet instance.
func (nic *NIC) Tx() (buf []byte) {
	var pkt *stack.PacketBuffer

	if pkt = nic.Link.Read(); pkt == nil {
		return
	}

	proto := make([]byte, 2)
	binary.BigEndian.PutUint16(proto, uint16(pkt.NetworkProtocolNumber))

	// Ethernet frame header
	buf = append(buf, pkt.EgressRoute.RemoteLinkAddress...)
	buf = append(buf, nic.mac...)
	buf = append(buf, proto...)

	for _, v := range pkt.AsSlices() {
		buf = append(buf, v...)
	}

	return
}
