// Copyright 2017 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package msgdepack is a simple event-driven algorith for parsing MsgPack documents
package msgdepack

import (
	"encoding/binary"
	"io"
	"math"
)

const (
	// Array indicates a MsgPack array type and PayloadInt64 is the array len
	Array = iota
	// Bool indicates a MsgPack boolean type and PayloadBool is valid
	Bool
	// Bytes indicates a MsgPack binary type and PayloadBytes is valid
	Bytes
	// EOF indicates that this MsgPack stream has ended and no payload is valid
	EOF
	// Error indicates that this MsgPack stream has an error and no payload is valid
	Error
	// Extension indicates a MsgPack extension type (first byte is extension type) and PayloadBytes is valid
	Extension
	// Float64 indicates a MsgPack float type and PayloadFloat64 is valid
	Float64
	// Int64 indicates a MsgPack numeric (except uint64) type and PayloadInt64 is valid
	Int64
	// Map indicates a MsgPack map type and PayloadInt64 is the key count
	Map
	// Nil indicates a MsgPack Nil type and no payload is valid
	Nil
	// String indicates a MsgPack string type and PayloadString is valid
	String
	// Uint64 indicates a MsgPack uint64 type and PayloadUint64 is valid
	Uint64
)

// MsgDepack is a stream based MsgPack decoder.
type MsgDepack struct {
	rs             io.ReadSeeker
	nextCalled     bool
	done           bool
	token          [1]byte
	elementsLeft   int64
	elementType    int
	buf            [8]byte
	payloadLen     int64
	payloadBool    bool
	payloadFloat64 float64
	payloadInt64   int64
	payloadString  string
	payloadUint64  uint64

	// The payload of the present token if it is of type Bool.  Only available
	PayloadBool *bool

	// The payload of the present token if it is of type Bytes
	PayloadBytes []byte

	// The payload of the present token if it is of type Float64
	PayloadFloat64 *float64

	// The payload of the present token if it is of type Int64
	PayloadInt64 *int64

	// The payload of the present token if it is of type String
	PayloadString *string

	// The payload of the present token if it is of type Uint64
	PayloadUint64 *uint64
}

// NewMsgDepack returns a new MsgDepack object for the specified io.ReadSeeker
func NewMsgDepack(rs io.ReadSeeker) *MsgDepack {
	makeArray()

	m := &MsgDepack{}
	m.rs = rs
	m.elementType = Error // Start as an error until the first element is reached
	m.elementsLeft = 1    // There must be at least 1 object

	return m
}

// Next reads the next MsgPack element and returns the element type, EOF if
// at the end of the stream, or Error if there was an error
func (m *MsgDepack) Next() int {
	m.nextCalled = true
	if !m.done {
		if 0 < m.elementsLeft {
			if _, err := m.rs.Seek(m.payloadLen, io.SeekCurrent); nil != err {
				goto fail
			} else {

				m.elementsLeft--
				m.payloadLen = 0
				m.PayloadBool = nil
				m.PayloadBytes = nil
				m.PayloadFloat64 = nil
				m.PayloadInt64 = nil
				m.PayloadString = nil
				m.PayloadUint64 = nil

				if _, err := io.ReadFull(m.rs, m.token[:]); nil != err {
					goto fail
				} else {
					header := msgDepackMap[m.token[0]].headerSize
					if _, err := io.ReadFull(m.rs, m.buf[:header]); nil != err {
						goto fail
					} else {
						msgDepackMap[m.token[0]].decodeHeader(m)
					}
				}
			}
		} else {
			m.done = true
			m.elementType = EOF
			m.cleanup()
		}
	}

	return m.elementType

fail:
	m.setError()
	return m.elementType
}

// Data populates the data for the token.  This is not automatically done
// to increase performance by skipping unwanted payloads.
func (m *MsgDepack) Data() bool {
	if (Error == m.elementType) || (EOF == m.elementType) {
		return false
	}

	if 0 < m.payloadLen {
		m.PayloadBytes = make([]byte, m.payloadLen)
		if _, err := io.ReadFull(m.rs, m.PayloadBytes); nil != err {
			m.setError()
			return false
		}
		m.payloadLen = 0
	}
	msgDepackMap[m.token[0]].decodePayload(m)
	return true
}

// setError puts the MsgDepack object into the error state.  This is not
// recoverable.
func (m *MsgDepack) setError() {
	m.elementType = Error
	m.done = true
	m.cleanup()
}

// cleanup sets the outward facing Payloadxxx and internal payloadxxx variables
// to nil or an empty set.
func (m *MsgDepack) cleanup() {
	m.elementsLeft = 0
	m.payloadLen = 0
	m.PayloadBool = nil
	m.PayloadBytes = nil
	m.PayloadFloat64 = nil
	m.PayloadInt64 = nil
	m.PayloadString = nil
	m.payloadString = ""
	m.PayloadUint64 = nil
}

func decodeHeaderPositiveFixInt(m *MsgDepack) {
	m.elementType = Int64
	m.payloadInt64 = int64(0x7f & m.token[0])
	m.PayloadInt64 = &m.payloadInt64
}

func decodeHeaderNil(m *MsgDepack) {
	m.elementType = Nil
}

func decodeHeaderError(m *MsgDepack) {
	m.setError()
}

func decodeHeaderFalse(m *MsgDepack) {
	m.elementType = Bool
	m.payloadBool = false
	m.PayloadBool = &m.payloadBool
}

func decodeHeaderTrue(m *MsgDepack) {
	m.elementType = Bool
	m.payloadBool = true
	m.PayloadBool = &m.payloadBool
}

func decodeHeaderNegativeFixInt(m *MsgDepack) {
	m.elementType = Int64
	m.payloadInt64 = int64(0 - (0x7f & m.token[0]))
	m.PayloadInt64 = &m.payloadInt64
}

func decodeHeaderUint8(m *MsgDepack) {
	m.elementType = Int64
	m.payloadLen = 1
}

func decodeHeaderUint16(m *MsgDepack) {
	m.elementType = Int64
	m.payloadLen = 2
}

func decodeHeaderUint32(m *MsgDepack) {
	m.elementType = Int64
	m.payloadLen = 4
}

func decodeHeaderUint64(m *MsgDepack) {
	m.elementType = Uint64
	m.payloadLen = 8
}

func decodeHeaderInt8(m *MsgDepack) {
	m.elementType = Int64
	m.payloadLen = 1
}

func decodeHeaderInt16(m *MsgDepack) {
	m.elementType = Int64
	m.payloadLen = 2
}

func decodeHeaderInt32(m *MsgDepack) {
	m.elementType = Int64
	m.payloadLen = 4
}

func decodeHeaderInt64(m *MsgDepack) {
	m.elementType = Int64
	m.payloadLen = 8
}

func decodeHeaderFloat32(m *MsgDepack) {
	m.elementType = Float64
	m.payloadLen = 4
}

func decodeHeaderFloat64(m *MsgDepack) {
	m.elementType = Float64
	m.payloadLen = 8
}

func decodeHeaderFixStr(m *MsgDepack) {
	m.elementType = String
	m.payloadLen = int64(0x1f & m.token[0])
}

func decodeHeaderStr8(m *MsgDepack) {
	m.elementType = String
	m.payloadLen = int64(uint8(m.buf[0]))
}

func decodeHeaderStr16(m *MsgDepack) {
	m.elementType = String
	m.payloadLen = int64(binary.BigEndian.Uint16(m.buf[:]))
}

func decodeHeaderStr32(m *MsgDepack) {
	m.elementType = String
	m.payloadLen = int64(binary.BigEndian.Uint32(m.buf[:]))
}

func decodeHeaderBin8(m *MsgDepack) {
	m.elementType = Bytes
	m.payloadLen = int64(uint8(m.buf[0]))
}

func decodeHeaderBin16(m *MsgDepack) {
	m.elementType = Bytes
	m.payloadLen = int64(binary.BigEndian.Uint16(m.buf[:]))
}

func decodeHeaderBin32(m *MsgDepack) {
	m.elementType = Bytes
	m.payloadLen = int64(binary.BigEndian.Uint32(m.buf[:]))
}

func decodeHeaderExt8(m *MsgDepack) {
	m.elementType = Extension
	m.payloadLen = int64(uint8(m.buf[0])) + 1
}

func decodeHeaderExt16(m *MsgDepack) {
	m.elementType = Extension
	m.payloadLen = int64(binary.BigEndian.Uint16(m.buf[:])) + 1
}

func decodeHeaderExt32(m *MsgDepack) {
	m.elementType = Extension
	m.payloadLen = int64(binary.BigEndian.Uint32(m.buf[:])) + 1
}

func decodeHeaderFixExt1(m *MsgDepack) {
	m.elementType = Extension
	m.payloadLen = 2
}

func decodeHeaderFixExt2(m *MsgDepack) {
	m.elementType = Extension
	m.payloadLen = 3
}

func decodeHeaderFixExt4(m *MsgDepack) {
	m.elementType = Extension
	m.payloadLen = 5
}

func decodeHeaderFixExt8(m *MsgDepack) {
	m.elementType = Extension
	m.payloadLen = 9
}

func decodeHeaderFixExt16(m *MsgDepack) {
	m.elementType = Extension
	m.payloadLen = 17
}

func decodeHeaderFixMap(m *MsgDepack) {
	m.elementType = Map
	m.payloadInt64 = int64(0x0f & m.token[0])
	m.PayloadInt64 = &m.payloadInt64
	m.elementsLeft += (m.payloadInt64 << 1)
}

func decodeHeaderMap16(m *MsgDepack) {
	m.elementType = Map
	m.payloadInt64 = int64(binary.BigEndian.Uint16(m.buf[:]))
	m.PayloadInt64 = &m.payloadInt64
	m.elementsLeft += (m.payloadInt64 << 1)
}

func decodeHeaderMap32(m *MsgDepack) {
	m.elementType = Map
	m.payloadInt64 = int64(binary.BigEndian.Uint32(m.buf[:]))
	m.PayloadInt64 = &m.payloadInt64
	m.elementsLeft += (m.payloadInt64 << 1)
}

func decodeHeaderFixArray(m *MsgDepack) {
	m.elementType = Array
	m.payloadInt64 = int64(0x0f & m.token[0])
	m.PayloadInt64 = &m.payloadInt64
	m.elementsLeft += m.payloadInt64
}

func decodeHeaderArray16(m *MsgDepack) {
	m.elementType = Array
	m.payloadInt64 = int64(binary.BigEndian.Uint16(m.buf[:]))
	m.PayloadInt64 = &m.payloadInt64
	m.elementsLeft += m.payloadInt64
}

func decodeHeaderArray32(m *MsgDepack) {
	m.elementType = Array
	m.payloadInt64 = int64(binary.BigEndian.Uint32(m.buf[:]))
	m.PayloadInt64 = &m.payloadInt64
	m.elementsLeft += m.payloadInt64
}

func decodePayloadEmpty(m *MsgDepack) {}

func decodePayloadString(m *MsgDepack) {
	m.payloadString = string(m.PayloadBytes)
	m.PayloadString = &m.payloadString
}

func decodePayloadFloat32(m *MsgDepack) {
	bits := binary.BigEndian.Uint32(m.PayloadBytes)
	m.payloadFloat64 = float64(math.Float32frombits(bits))
	m.PayloadFloat64 = &m.payloadFloat64
}

func decodePayloadFloat64(m *MsgDepack) {
	bits := binary.BigEndian.Uint64(m.PayloadBytes)
	m.payloadFloat64 = math.Float64frombits(bits)
	m.PayloadFloat64 = &m.payloadFloat64
}

func decodePayloadUint8(m *MsgDepack) {
	m.payloadInt64 = int64(uint8(m.PayloadBytes[0]))
	m.PayloadInt64 = &m.payloadInt64
}

func decodePayloadInt8(m *MsgDepack) {
	m.payloadInt64 = int64(int8(m.PayloadBytes[0]))
	m.PayloadInt64 = &m.payloadInt64
}

func decodePayloadUint16(m *MsgDepack) {
	m.payloadInt64 = int64(binary.BigEndian.Uint16(m.PayloadBytes))
	m.PayloadInt64 = &m.payloadInt64
}

func decodePayloadInt16(m *MsgDepack) {
	_ = m.PayloadBytes[1] // bounds check hint to compiler; see golang.org/issue/14808
	m.payloadInt64 = int64(int16(m.PayloadBytes[0])<<8 | int16(m.PayloadBytes[1]))
	m.PayloadInt64 = &m.payloadInt64
}

func decodePayloadUint32(m *MsgDepack) {
	m.payloadInt64 = int64(binary.BigEndian.Uint32(m.PayloadBytes))
	m.PayloadInt64 = &m.payloadInt64
}

func decodePayloadInt32(m *MsgDepack) {
	_ = m.PayloadBytes[3] // bounds check hint to compiler; see golang.org/issue/14808
	m.payloadInt64 = int64(int32(m.PayloadBytes[0])<<24 | int32(m.PayloadBytes[1])<<16 | int32(m.PayloadBytes[2])<<8 | int32(m.PayloadBytes[3]))
	m.PayloadInt64 = &m.payloadInt64
}

func decodePayloadUint64(m *MsgDepack) {
	m.payloadUint64 = binary.BigEndian.Uint64(m.PayloadBytes)
	m.PayloadUint64 = &m.payloadUint64
}

func decodePayloadInt64(m *MsgDepack) {
	_ = m.PayloadBytes[7] // bounds check hint to compiler; see golang.org/issue/14808
	m.payloadInt64 = int64(m.PayloadBytes[0])<<56 | int64(m.PayloadBytes[1])<<48 | int64(m.PayloadBytes[2])<<40 | int64(m.PayloadBytes[3])<<32 | int64(m.PayloadBytes[4])<<24 | int64(m.PayloadBytes[5])<<16 | int64(m.PayloadBytes[6])<<8 | int64(m.PayloadBytes[7])
	m.PayloadInt64 = &m.payloadInt64
}

type msgDepackCtrl struct {
	headerSize    int64
	decodeHeader  func(*MsgDepack)
	decodePayload func(*MsgDepack)
}

var msgDepackMap []msgDepackCtrl

/*
var msgPackMap = [256]msgDepackCtrl{
	{decodeHeader: decodeHeaderPositiveFixInt, decodePayload: decodePayloadEmpty},
}
*/

func makeArray() {
	tmp := make([]msgDepackCtrl, 256)

	// Set the default payload decoder
	for i := 0; i < 0xff; i++ {
		tmp[i].decodePayload = decodePayloadEmpty
	}

	for i := 0; i < 0x80; i++ {
		tmp[i].decodeHeader = decodeHeaderPositiveFixInt
	}
	for i := 0x80; i < 0x90; i++ {
		tmp[i].decodeHeader = decodeHeaderFixMap
	}
	for i := 0x90; i < 0xa0; i++ {
		tmp[i].decodeHeader = decodeHeaderFixArray
	}
	for i := 0xa0; i < 0xc0; i++ {
		tmp[i].decodeHeader = decodeHeaderFixStr
		tmp[i].decodePayload = decodePayloadString
	}

	tmp[0xc0].decodeHeader = decodeHeaderNil

	tmp[0xc1].decodeHeader = decodeHeaderError

	tmp[0xc2].decodeHeader = decodeHeaderFalse

	tmp[0xc3].decodeHeader = decodeHeaderTrue

	tmp[0xc4].headerSize = 1
	tmp[0xc4].decodeHeader = decodeHeaderBin8

	tmp[0xc5].headerSize = 2
	tmp[0xc5].decodeHeader = decodeHeaderBin16

	tmp[0xc6].headerSize = 4
	tmp[0xc6].decodeHeader = decodeHeaderBin32

	tmp[0xc7].headerSize = 1
	tmp[0xc7].decodeHeader = decodeHeaderExt8

	tmp[0xc8].headerSize = 2
	tmp[0xc8].decodeHeader = decodeHeaderExt16

	tmp[0xc9].headerSize = 4
	tmp[0xc9].decodeHeader = decodeHeaderExt32

	tmp[0xca].decodeHeader = decodeHeaderFloat32
	tmp[0xca].decodePayload = decodePayloadFloat32

	tmp[0xcb].decodeHeader = decodeHeaderFloat64
	tmp[0xcb].decodePayload = decodePayloadFloat64

	tmp[0xcc].decodeHeader = decodeHeaderUint8
	tmp[0xcc].decodePayload = decodePayloadUint8

	tmp[0xcd].decodeHeader = decodeHeaderUint16
	tmp[0xcd].decodePayload = decodePayloadUint16

	tmp[0xce].decodeHeader = decodeHeaderUint32
	tmp[0xce].decodePayload = decodePayloadUint32

	tmp[0xcf].decodeHeader = decodeHeaderUint64
	tmp[0xcf].decodePayload = decodePayloadUint64

	tmp[0xd0].decodeHeader = decodeHeaderInt8
	tmp[0xd0].decodePayload = decodePayloadInt8

	tmp[0xd1].decodeHeader = decodeHeaderInt16
	tmp[0xd1].decodePayload = decodePayloadInt16

	tmp[0xd2].decodeHeader = decodeHeaderInt32
	tmp[0xd2].decodePayload = decodePayloadInt32

	tmp[0xd3].decodeHeader = decodeHeaderInt64
	tmp[0xd3].decodePayload = decodePayloadInt64

	tmp[0xd4].decodeHeader = decodeHeaderFixExt1

	tmp[0xd5].decodeHeader = decodeHeaderFixExt2

	tmp[0xd6].decodeHeader = decodeHeaderFixExt4

	tmp[0xd7].decodeHeader = decodeHeaderFixExt8

	tmp[0xd8].decodeHeader = decodeHeaderFixExt16

	tmp[0xd9].headerSize = 1
	tmp[0xd9].decodeHeader = decodeHeaderStr8
	tmp[0xd9].decodePayload = decodePayloadString

	tmp[0xda].headerSize = 2
	tmp[0xda].decodeHeader = decodeHeaderStr16
	tmp[0xda].decodePayload = decodePayloadString

	tmp[0xdb].headerSize = 4
	tmp[0xdb].decodeHeader = decodeHeaderStr32
	tmp[0xdb].decodePayload = decodePayloadString

	tmp[0xdc].headerSize = 2
	tmp[0xdc].decodeHeader = decodeHeaderArray16

	tmp[0xdd].headerSize = 4
	tmp[0xdd].decodeHeader = decodeHeaderArray32

	tmp[0xde].headerSize = 2
	tmp[0xde].decodeHeader = decodeHeaderMap16

	tmp[0xdf].headerSize = 4
	tmp[0xdf].decodeHeader = decodeHeaderMap32

	for i := 0xe0; i < 0x100; i++ {
		tmp[i].decodeHeader = decodeHeaderNegativeFixInt
	}

	msgDepackMap = tmp
}
