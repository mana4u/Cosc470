package main

import (
	"errors"
	"fmt"
	"net"
	"runtime"
	"syscall"
	"unsafe"
	"log"
)

var (
	errMissingAddress  = errors.New("missing address")
	errMissingHeader   = errors.New("missing header")
	errHeaderTooShort  = errors.New("header too short")
	errBufferTooShort  = errors.New("buffer too short")
	errInvalidConnType = errors.New("invalid conn type")
)

// References:
//
// RFC  791  Internet Protocol
//	http://tools.ietf.org/html/rfc791
// RFC 1112  Host Extensions for IP Multicasting
//	http://tools.ietf.org/html/rfc1112
// RFC 1122  Requirements for Internet Hosts
//	http://tools.ietf.org/html/rfc1122

const (
	Version      = 4  // protocol version
	HeaderLen    = 20 // header length without extension headers
	maxHeaderLen = 60 // sensible default, revisit if later RFCs define new usage of version and header length fields
)

const (
	posTOS      = 1  // type-of-service
	posTotalLen = 2  // packet total length
	posID       = 4  // identification
	posFragOff  = 6  // fragment offset
	posTTL      = 8  // time-to-live
	posProtocol = 9  // next protocol
	posChecksum = 10 // checksum
	posSrc      = 12 // source address
	posDst      = 16 // destination address
)

type HeaderFlags int

const (
	MoreFragments HeaderFlags = 1 << iota // more fragments flag
	DontFragment                          // don't fragment flag
)

// A Header represents an IPv4 header.
type Header struct {
	Version  int         // protocol version
	Len      int         // header length
	TOS      int         // type-of-service
	TotalLen int         // packet total length
	ID       int         // identification
	Flags    HeaderFlags // flags
	FragOff  int         // fragment offset
	TTL      int         // time-to-live
	Protocol int         // next protocol
	Checksum int         // checksum
	Src      net.IP      // source address
	Dst      net.IP      // destination address
	Options  []byte      // options, extension headers
}

func (h *Header) String() string {
	if h == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ver: %v, hdrlen: %v, tos: %#x, totallen: %v, id: %#x, flags: %#x, fragoff: %#x, ttl: %v, proto: %v, cksum: %#x, src: %v, dst: %v", h.Version, h.Len, h.TOS, h.TotalLen, h.ID, h.Flags, h.FragOff, h.TTL, h.Protocol, h.Checksum, h.Src, h.Dst)
}

// Please refer to the online manual; IP(4) on Darwin, FreeBSD and
// OpenBSD.  IP(7) on Linux.
const supportsNewIPInput = runtime.GOOS == "linux" || runtime.GOOS == "openbsd"

// Marshal returns the binary encoding of the IPv4 header h.
func (h *Header) Marshal() ([]byte, error) {
	if h == nil {
		return nil, syscall.EINVAL
	}
	if h.Len < HeaderLen {
		return nil, errHeaderTooShort
	}
	hdrlen := HeaderLen + len(h.Options)
	b := make([]byte, hdrlen)
	b[0] = byte(Version<<4 | (hdrlen >> 2 & 0x0f))
	b[posTOS] = byte(h.TOS)
	flagsAndFragOff := (h.FragOff & 0x1fff) | int(h.Flags<<13)
	if supportsNewIPInput {
		b[posTotalLen], b[posTotalLen+1] = byte(h.TotalLen>>8), byte(h.TotalLen)
		b[posFragOff], b[posFragOff+1] = byte(flagsAndFragOff>>8), byte(flagsAndFragOff)
	} else {
		*(*uint16)(unsafe.Pointer(&b[posTotalLen : posTotalLen+1][0])) = uint16(h.TotalLen)
		*(*uint16)(unsafe.Pointer(&b[posFragOff : posFragOff+1][0])) = uint16(flagsAndFragOff)
	}
	b[posID], b[posID+1] = byte(h.ID>>8), byte(h.ID)
	b[posTTL] = byte(h.TTL)
	b[posProtocol] = byte(h.Protocol)
	b[posChecksum], b[posChecksum+1] = byte(h.Checksum>>8), byte(h.Checksum)
	if ip := h.Src.To4(); ip != nil {
		copy(b[posSrc:posSrc+net.IPv4len], ip[:net.IPv4len])
	}
	if ip := h.Dst.To4(); ip != nil {
		copy(b[posDst:posDst+net.IPv4len], ip[:net.IPv4len])
	} else {
		return nil, errMissingAddress
	}
	if len(h.Options) > 0 {
		copy(b[HeaderLen:], h.Options)
	}
	return b, nil
}

// ParseHeader parses b as an IPv4 header.
func ParseHeader(b []byte) (*Header, error) {
	if len(b) < HeaderLen {
		return nil, errHeaderTooShort
	}
	hdrlen := int(b[0]&0x0f) << 2
	if hdrlen > len(b) {
		return nil, errBufferTooShort
	}
	h := &Header{}
	h.Version = int(b[0] >> 4)
	h.Len = hdrlen
	h.TOS = int(b[posTOS])
	if supportsNewIPInput {
		h.TotalLen = int(b[posTotalLen])<<8 | int(b[posTotalLen+1])
		h.FragOff = int(b[posFragOff])<<8 | int(b[posFragOff+1])
	} else {
		h.TotalLen = int(*(*uint16)(unsafe.Pointer(&b[posTotalLen : posTotalLen+1][0])))
		h.TotalLen += hdrlen
		h.FragOff = int(*(*uint16)(unsafe.Pointer(&b[posFragOff : posFragOff+1][0])))
	}
	h.Flags = HeaderFlags(h.FragOff&0xe000) >> 13
	h.FragOff = h.FragOff & 0x1fff
	h.ID = int(b[posID])<<8 | int(b[posID+1])
	h.TTL = int(b[posTTL])
	h.Protocol = int(b[posProtocol])
	h.Checksum = int(b[posChecksum])<<8 | int(b[posChecksum+1])
	h.Src = net.IPv4(b[posSrc], b[posSrc+1], b[posSrc+2], b[posSrc+3])
	h.Dst = net.IPv4(b[posDst], b[posDst+1], b[posDst+2], b[posDst+3])
	if hdrlen-HeaderLen > 0 {
		h.Options = make([]byte, hdrlen-HeaderLen)
		copy(h.Options, b[HeaderLen:])
	}
	return h, nil
}


func main() {
	fd, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	addr := syscall.SockaddrInet4{
		Port: 0,
		Addr: [4]byte{127, 0, 0, 1},
	}

  	fmt.Println("Enter tcp, udp, or icmp")
	var input string
	fmt.Scanln(&input)
	
	switch input {
	case "tcp":
		var err error
		h := Header{
			Version:  4,
			Len:      20,
			TotalLen: 20, // 20 bytes for IP, 10 for ICMP
			TTL:      64,
			Protocol: 6, // TCP
			Dst:      net.IPv4(127, 0, 0, 1),
			// ID, Src and Checksum will be set for us by the kernel
		}
		out, err := h.Marshal()
		if err != nil {
			log.Fatal(err)
		}
		err = syscall.Sendto(fd, out, 0, &addr)
		if err != nil {
			log.Fatal("Sendto:", err)
		}
	  case "udp":
		var err error
		h := Header{
			Version:  4,
			Len:      20,
			TotalLen: 20, // 20 bytes for IP, 10 for ICMP
			TTL:      64,
			Protocol: 17, // UDP
			Dst:      net.IPv4(127, 0, 0, 1),
			// ID, Src and Checksum will be set for us by the kernel
		}
		out, err := h.Marshal()
		if err != nil {
			log.Fatal(err)
		}
		err = syscall.Sendto(fd, out, 0, &addr)
		if err != nil {
			log.Fatal("Sendto:", err)
		}
	case "icmp":
		var err error
		h := Header{
			Version:  4,
			Len:      20,
			TotalLen: 20, // 20 bytes for IP, 10 for ICMP
			TTL:      64,
			Protocol: 1, // ICMP
			Dst:      net.IPv4(127, 0, 0, 1),
			// ID, Src and Checksum will be set for us by the kernel
		}

		icmp := []byte{
			8, // type: echo request
			0, // code: not used by echo request
			0, // checksum (16 bit), we fill in below
			0,
			0, // identifier (16 bit). zero allowed.
			0,
			0, // sequence number (16 bit). zero allowed.
			0,
			0xC0, // Optional data. ping puts time packet sent here
			0xDE,
		}
		cs := csum(icmp)
		icmp[2] = byte(cs)
		icmp[3] = byte(cs >> 8)

		out, err := h.Marshal()
		if err != nil {
			log.Fatal(err)
		}
		ic := append(out, icmp...)

		err = syscall.Sendto(fd, ic, 0, &addr)
		if err != nil {
			log.Fatal("Sendto:", err)
		}
		default: break
	}

}
func csum(b []byte) uint16 {
	var s uint32
	for i := 0; i < len(b); i += 2 {
		s += uint32(b[i+1])<<8 | uint32(b[i])
	}
	// add back the carry
	s = s>>16 + s&0xffff
	s = s + s>>16
	return uint16(^s)
}

