package wol

////////////////////////////////////////////////////////////////////////////////

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"regexp"

	"github.com/jonathongardner/fife/logger"
	"github.com/sirupsen/logrus"
)

////////////////////////////////////////////////////////////////////////////////

var (
	delims = ":-"
	reMAC  = regexp.MustCompile(`^([0-9a-fA-F]{2}[` + delims + `]){5}([0-9a-fA-F]{2})$`)
)

////////////////////////////////////////////////////////////////////////////////

// MACAddress represents a 6 byte network mac address.
type MACAddress [6]byte

// MagicPacket is constituted of 6 bytes of 0xFF followed by 16-groups of the
// destination MAC address.
type MagicPacket struct {
	header  [6]byte
	payload [16]MACAddress
}

// New returns a magic packet based on a mac address string.
func New(mac string) (*MagicPacket, error) {
	var packet MagicPacket
	var macAddr MACAddress

	hwAddr, err := net.ParseMAC(mac)
	if err != nil {
		return nil, err
	}

	// We only support 6 byte MAC addresses since it is much harder to use the
	// binary.Write(...) interface when the size of the MagicPacket is dynamic.
	if !reMAC.MatchString(mac) {
		return nil, fmt.Errorf("%s is not a IEEE 802 MAC-48 address", mac)
	}

	// Copy bytes from the returned HardwareAddr -> a fixed size MACAddress.
	for idx := range macAddr {
		macAddr[idx] = hwAddr[idx]
	}

	// Setup the header which is 6 repetitions of 0xFF.
	for idx := range packet.header {
		packet.header[idx] = 0xFF
	}

	// Setup the payload which is 16 repetitions of the MAC addr.
	for idx := range packet.payload {
		packet.payload[idx] = macAddr
	}

	return &packet, nil
}

// Marshal serializes the magic packet structure into a 102 byte slice.
func (mp *MagicPacket) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.BigEndian, mp); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func Run(ctx context.Context, macAddr string) (err error) {
	lgr := logger.ContextLogger(ctx)

	var localAddr *net.UDPAddr

	bcastAddr := fmt.Sprintf("%s:%s", "255.255.255.255", "9")
	udpAddr, err := net.ResolveUDPAddr("udp", bcastAddr)
	if err != nil {
		return fmt.Errorf("error resolving udp address: %w", err)
	}

	// Build the magic packet.
	mp, err := New(macAddr)
	if err != nil {
		return fmt.Errorf("error creating magic packet: %w", err)
	}

	// Grab a stream of bytes to send.
	bs, err := mp.Marshal()
	if err != nil {
		return fmt.Errorf("error marshaling magic packet: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error (1): %w", err)
	}

	// Grab a UDP connection to send our packet of bytes.
	conn, err := net.DialUDP("udp", localAddr, udpAddr)
	if err != nil {
		return fmt.Errorf("error dialing udp: %w", err)
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("error closing connection: %w", cerr))
		}
	}()

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error (2): %w", err)
	}

	n, err := conn.Write(bs)
	if err == nil && n != 102 {
		err = fmt.Errorf("magic packet sent was %d bytes (expected 102 bytes sent)", n)
	}
	lgr.WithFields(
		logrus.Fields{
			"mac":            macAddr,
			"broadcastingTo": bcastAddr,
			"error":          err,
		},
	).Infof("Sent magic packet")
	if err != nil {
		return fmt.Errorf("error sending magic packet: %w", err)
	}

	return nil
}
