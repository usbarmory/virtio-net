VirtIO Network driver
=====================

This Go package implements TCP/IP connectivity through VirtIO Networking to be
used with `GOOS=tamago` as supported by the
[TamaGo](https://github.com/usbarmory/tamago) framework for bare metal Go on
AMD64/ARM/RISC-V processors.

The package supports TCP/IP networking through gVisor (`go` branch)
[tcpip](https://pkg.go.dev/gvisor.dev/gvisor/pkg/tcpip)
stack pure Go implementation.

The interface TCP/IP stack can be attached to the Go runtime by setting
`net.SocketFunc` to the interface `Socket` function:

```
# VirtIO over MMIO
dev := &vnet.Net{
	Transport: &virtio.MMIO{
		Base: vm.VIRTIO_NET0_BASE,
	}
}

# VirtIO over PCI
dev := &vnet.Net{
	Transport: &virtio.PCI{
		Device: pci.Probe(
			0,
			vm.VIRTIO_NET_PCI_VENDOR,
			vm.VIRTIO_NET_PCI_DEVICE,
		),
	}
}

iface, _ := vnet.Init(dev, "10.0.0.1", "255.255.255.0", "10.0.0.2")

net.SocketFunc = iface.Socket
```

See [tamago-example](https://github.com/usbarmory/tamago-example/blob/master/network/virtio.go)
for a full integration example.

Authors
=======

Andrea Barisani  
andrea@inversepath.com  

Andrej Rosano  
andrej@inversepath.com  

Documentation
=============

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/virtio-net).


For more information about TamaGo see its
[repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki).

License
=======

tamago | https://github.com/usbarmory/virtio-net  
Copyright (c) WithSecure Corporation

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/virtio-net/blob/master/LICENSE) file.
