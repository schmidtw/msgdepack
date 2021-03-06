package msgdepack

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func getStringPtr(s string) *string { return &s }

func TestSingleElement(t *testing.T) {
	type test struct {
		element        int
		in             []byte
		payloadBool    *bool
		payloadBytes   []byte
		payloadFloat64 *float64
		payloadInt64   *int64
		payloadString  *string
		payloadUint64  *uint64
	}

	False := false
	True := true
	//Float64 := float64(3.2)

	simpleTests := []test{
		// Small fixed values
		{element: Int64, in: []byte{0x00}, payloadInt64: &[]int64{0}[0]},
		{element: Int64, in: []byte{0x01}, payloadInt64: &[]int64{1}[0]},
		{element: Int64, in: []byte{0x7f}, payloadInt64: &[]int64{127}[0]},
		{element: Map, in: []byte{0x80}, payloadInt64: &[]int64{0}[0]},
		{element: Array, in: []byte{0x90}, payloadInt64: &[]int64{0}[0]},
		{element: String, in: []byte{0xa0}, payloadString: getStringPtr("")},
		{element: String, in: []byte{0xa3, 'c', 'a', 't'}, payloadBytes: []byte{'c', 'a', 't'}, payloadString: getStringPtr("cat")},
		{element: Int64, in: []byte{0xe0}, payloadInt64: &[]int64{-32}[0]},
		{element: Int64, in: []byte{0xe1}, payloadInt64: &[]int64{-31}[0]},

		// Nil
		{element: Nil, in: []byte{0xc0}},

		// Error
		{element: Error, in: []byte{0xc1}},

		// Bool
		{element: Bool, in: []byte{0xc2}, payloadBool: &False},
		{element: Bool, in: []byte{0xc3}, payloadBool: &True},

		// Bytes
		{element: Bytes, in: []byte{0xc4, 0}},
		{element: Bytes, in: []byte{0xc4, 1, 33}, payloadBytes: []byte{33}},
		{element: Bytes, in: []byte{0xc5, 0, 0}},
		{element: Bytes, in: []byte{0xc5, 0, 1, 20}, payloadBytes: []byte{20}},
		{element: Bytes, in: []byte{0xc6, 0, 0, 0, 0}},
		{element: Bytes, in: []byte{0xc6, 0, 0, 0, 1, 40}, payloadBytes: []byte{40}},

		// Extensions
		{element: Extension, in: []byte{0xc7, 0, 10}, payloadBytes: []byte{10}},
		{element: Extension, in: []byte{0xc7, 1, 10, 12}, payloadBytes: []byte{10, 12}},
		{element: Extension, in: []byte{0xc8, 0, 0, 11}, payloadBytes: []byte{11}},
		{element: Extension, in: []byte{0xc8, 0, 1, 14, 13}, payloadBytes: []byte{14, 13}},
		{element: Extension, in: []byte{0xc9, 0, 0, 0, 0, 21}, payloadBytes: []byte{21}},
		{element: Extension, in: []byte{0xc9, 0, 0, 0, 1, 21, 22}, payloadBytes: []byte{21, 22}},

		// Floats
		{element: Float64, in: []byte{0xca, 0x40, 0x09, 0x99, 0x99}, payloadFloat64: &[]float64{2.15}[0]},
		{element: Float64, in: []byte{0xcb, 0x40, 0x09, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9a}, payloadFloat64: &[]float64{3.2}[0]},

		// Integers (signed & unsigned)
		{element: Int64, in: []byte{0xcc, 0x7f}, payloadInt64: &[]int64{127}[0]},
		{element: Int64, in: []byte{0xcc, 0xff}, payloadInt64: &[]int64{255}[0]},
		{element: Int64, in: []byte{0xcd, 0x01, 0xff}, payloadInt64: &[]int64{511}[0]},
		{element: Int64, in: []byte{0xcd, 0xff, 0xff}, payloadInt64: &[]int64{65535}[0]},
		{element: Int64, in: []byte{0xce, 0, 0, 0x01, 0xff}, payloadInt64: &[]int64{0x1ff}[0]},
		{element: Int64, in: []byte{0xce, 0xff, 0xff, 0xff, 0xff}, payloadInt64: &[]int64{0xffffffff}[0]},
		{element: Uint64, in: []byte{0xcf, 0, 0, 0, 0, 0, 0, 0x01, 0xff}, payloadUint64: &[]uint64{0x1ff}[0]},
		{element: Uint64, in: []byte{0xcf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, payloadUint64: &[]uint64{0xffffffffffffffff}[0]},
		{element: Int64, in: []byte{0xd0, 0x7f}, payloadInt64: &[]int64{127}[0]},
		{element: Int64, in: []byte{0xd0, 0xff}, payloadInt64: &[]int64{-1}[0]},
		{element: Int64, in: []byte{0xd1, 0x01, 0xff}, payloadInt64: &[]int64{511}[0]},
		{element: Int64, in: []byte{0xd1, 0xff, 0xff}, payloadInt64: &[]int64{-1}[0]},
		{element: Int64, in: []byte{0xd2, 0, 0, 0x01, 0xff}, payloadInt64: &[]int64{0x1ff}[0]},
		{element: Int64, in: []byte{0xd2, 0xff, 0xff, 0xff, 0xff}, payloadInt64: &[]int64{-1}[0]},
		{element: Int64, in: []byte{0xd3, 0, 0, 0, 0, 0, 0, 0x01, 0xff}, payloadInt64: &[]int64{0x1ff}[0]},
		{element: Int64, in: []byte{0xd3, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, payloadInt64: &[]int64{-1}[0]},

		// Fixed extensions
		{element: Extension, in: []byte{0xd4, 12, 13}, payloadBytes: []byte{12, 13}},
		{element: Extension, in: []byte{0xd5, 12, 13, 14}, payloadBytes: []byte{12, 13, 14}},
		{element: Extension, in: []byte{0xd6, 12, 13, 14, 15, 16}, payloadBytes: []byte{12, 13, 14, 15, 16}},
		{element: Extension, in: []byte{0xd7, 12, 13, 14, 15, 16, 20, 21, 22, 23}, payloadBytes: []byte{12, 13, 14, 15, 16, 20, 21, 22, 23}},
		{element: Extension, in: []byte{0xd8, 12, 13, 14, 15, 16, 20, 21, 22, 23, 30, 31, 32, 33, 34, 35, 36, 37}, payloadBytes: []byte{12, 13, 14, 15, 16, 20, 21, 22, 23, 30, 31, 32, 33, 34, 35, 36, 37}},

		// Strings
		{element: String, in: []byte{0xd9, 0}, payloadString: getStringPtr("")},
		{element: String, in: []byte{0xd9, 3, 'c', 'a', 't'}, payloadBytes: []byte{'c', 'a', 't'}, payloadString: getStringPtr("cat")},
		{element: String, in: []byte{0xda, 0, 0}, payloadString: getStringPtr("")},
		{element: String, in: []byte{0xda, 0, 3, 'c', 'a', 't'}, payloadBytes: []byte{'c', 'a', 't'}, payloadString: getStringPtr("cat")},
		{element: String, in: []byte{0xdb, 0, 0, 0, 0}, payloadString: getStringPtr("")},
		{element: String, in: []byte{0xdb, 0, 0, 0, 3, 'c', 'a', 't'}, payloadBytes: []byte{'c', 'a', 't'}, payloadString: getStringPtr("cat")},

		// Arrays of length 0
		{element: Array, in: []byte{0xdc, 0, 0}, payloadInt64: &[]int64{0}[0]},
		{element: Array, in: []byte{0xdd, 0, 0, 0, 0}, payloadInt64: &[]int64{0}[0]},

		// Maps of length 0
		{element: Map, in: []byte{0xde, 0, 0}, payloadInt64: &[]int64{0}[0]},
		{element: Map, in: []byte{0xdf, 0, 0, 0, 0}, payloadInt64: &[]int64{0}[0]},
	}

	assert := assert.New(t)

	for _, e := range simpleTests {
		md := NewMsgDepack(bytes.NewReader(e.in))

		element := md.Next()
		assert.Equal(e.element, element, "Element types must match")

		assert.Equal((Error != element), md.Data(), "Data must be there.")
		// Repeated calls should be consistant
		assert.Equal((Error != element), md.Data(), "Data must be there.")
		assert.Equal((Error != element), md.Data(), "Data must be there.")

		// Check all types of payloads
		assert.Equal(e.payloadBool, md.PayloadBool, "PayloadBool should match")
		assert.Equal(e.payloadBytes, md.PayloadBytes, "PayloadBytes should match")
		assert.Equal(e.payloadInt64, md.PayloadInt64, "PayloadInt64 should match")
		assert.Equal(e.payloadString, md.PayloadString, "PayloadString should match")
		assert.Equal(e.payloadUint64, md.PayloadUint64, "PayloadUint64 should match")
		if nil != e.payloadFloat64 {
			assert.InEpsilon(*e.payloadFloat64, *md.PayloadFloat64, float64(0.00001), "PayloadFloat64 should match")
		} else {
			assert.Nil(md.PayloadFloat64, "PayloadFloat64 should be nil")
		}

		if Error == element {
			// validate the Error logic works correctly
			element = md.Next()
			assert.Equal(Error, element, "Next element should be Error")
			assert.Equal(false, md.Data(), "Data must not there.")
			element = md.Next()
			assert.Equal(Error, element, "Next element should be Error")
			assert.Equal(false, md.Data(), "Data must not there.")
		} else {
			// validate the EOF logic works correctly
			element = md.Next()
			assert.Equal(EOF, element, "Next element should be EOF")
			assert.Equal(false, md.Data(), "Data must not there.")
			element = md.Next()
			assert.Equal(EOF, element, "Next element should be EOF")
			assert.Equal(false, md.Data(), "Data must not there.")
		}
	}
}

func TestEarlyFailures(t *testing.T) {
	type test struct {
		elementNow  int
		elementNext int
		dataPresent bool
		in          []byte
	}

	/* Run a set of tests to check if missing data is tolerated. */
	errors := []test{
		{elementNow: Error, elementNext: Error, in: []byte{0xc4}},
		{elementNow: Error, elementNext: Error, in: []byte{0xc5}},
		{elementNow: Error, elementNext: Error, in: []byte{0xc6}},
		{elementNow: Error, elementNext: Error, in: []byte{0xc7}},
		{elementNow: Error, elementNext: Error, in: []byte{0xc8}},
		{elementNow: Error, elementNext: Error, in: []byte{0xc9}},
		{elementNow: Float64, elementNext: EOF, in: []byte{0xca}},
		{elementNow: Float64, elementNext: EOF, in: []byte{0xcb}},
		{elementNow: Int64, elementNext: EOF, in: []byte{0xcc}},
		{elementNow: Int64, elementNext: EOF, in: []byte{0xcd}},
		{elementNow: Int64, elementNext: EOF, in: []byte{0xce}},
		{elementNow: Uint64, elementNext: EOF, in: []byte{0xcf}},
		{elementNow: Int64, elementNext: EOF, in: []byte{0xd0}},
		{elementNow: Int64, elementNext: EOF, in: []byte{0xd1}},
		{elementNow: Int64, elementNext: EOF, in: []byte{0xd2}},
		{elementNow: Int64, elementNext: EOF, in: []byte{0xd3}},
		{elementNow: Extension, elementNext: EOF, in: []byte{0xd4}},
		{elementNow: Extension, elementNext: EOF, in: []byte{0xd5}},
		{elementNow: Extension, elementNext: EOF, in: []byte{0xd6}},
		{elementNow: Extension, elementNext: EOF, in: []byte{0xd7}},
		{elementNow: Extension, elementNext: EOF, in: []byte{0xd8}},
		{elementNow: Error, elementNext: Error, in: []byte{0xd9}},
		{elementNow: Error, elementNext: Error, in: []byte{0xda}},
		{elementNow: Error, elementNext: Error, in: []byte{0xdb}},
		{elementNow: Error, elementNext: Error, in: []byte{0xdc}},
		{elementNow: Error, elementNext: Error, in: []byte{0xdd}},
		{elementNow: Error, elementNext: Error, in: []byte{0xde}},
		{elementNow: Error, elementNext: Error, in: []byte{0xdf}},
		{elementNow: Map, dataPresent: true, elementNext: Error, in: []byte{0x81}},
		{elementNow: Array, dataPresent: true, elementNext: Error, in: []byte{0x91}},
		{elementNow: String, elementNext: EOF, in: []byte{0xa3}},
	}

	assert := assert.New(t)

	for _, e := range errors {
		md := NewMsgDepack(bytes.NewReader(e.in))

		element := md.Next()
		assert.Equal(e.elementNow, element, "Element types must match")

		assert.Equal(e.elementNext, md.Next(), "Next() element must be Error")
		assert.Equal(false, md.Data(), "Data() must be false")
		assert.Equal(e.elementNext, md.Next(), "Next() element must be Error, %v", e.in)
		assert.Equal(false, md.Data(), "Data() must be false")

		/* -- change the order of Next() and Data() -------------------- */

		md = NewMsgDepack(bytes.NewReader(e.in))

		element = md.Next()
		assert.Equal(e.elementNow, element, "Element types must match")

		assert.Equal(e.dataPresent, md.Data(), "Data() must be false")
		assert.Equal(Error, md.Next(), "Next() element must be Error")
		assert.Equal(false, md.Data(), "Data() must be false")
	}
}

func TestDecode(t *testing.T) {
	in := []byte{0x85, 0xa8, 0x6d, 0x73, 0x67, 0x5f, 0x74, 0x79,
		0x70, 0x65, 0x03, 0xb0, 0x74, 0x72, 0x61, 0x6e,
		0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x75, 0x75,
		0x69, 0x64, 0xd9, 0x24, 0x39, 0x34, 0x34, 0x37,
		0x32, 0x34, 0x31, 0x63, 0x2d, 0x35, 0x32, 0x33,
		0x38, 0x2d, 0x34, 0x63, 0x62, 0x39, 0x2d, 0x39,
		0x62, 0x61, 0x61, 0x2d, 0x37, 0x30, 0x37, 0x36,
		0x65, 0x33, 0x32, 0x33, 0x32, 0x38, 0x39, 0x39,
		0xa6, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0xd9,
		0x26, 0x64, 0x6e, 0x73, 0x3a, 0x77, 0x65, 0x62,
		0x70, 0x61, 0x2e, 0x63, 0x6f, 0x6d, 0x63, 0x61,
		0x73, 0x74, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76,
		0x32, 0x2d, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65,
		0x2d, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0xa4,
		0x64, 0x65, 0x73, 0x74, 0xb2, 0x73, 0x65, 0x72,
		0x69, 0x61, 0x6c, 0x3a, 0x31, 0x32, 0x33, 0x34,
		0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0xa7,
		0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0xc4,
		0x45, 0x7b, 0x20, 0x22, 0x6e, 0x61, 0x6d, 0x65,
		0x73, 0x22, 0x3a, 0x20, 0x5b, 0x20, 0x22, 0x44,
		0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x58, 0x5f,
		0x43, 0x49, 0x53, 0x43, 0x4f, 0x5f, 0x43, 0x4f,
		0x4d, 0x5f, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69,
		0x74, 0x79, 0x2e, 0x46, 0x69, 0x72, 0x65, 0x77,
		0x61, 0x6c, 0x6c, 0x2e, 0x46, 0x69, 0x72, 0x65,
		0x77, 0x61, 0x6c, 0x6c, 0x4c, 0x65, 0x76, 0x65,
		0x6c, 0x22, 0x20, 0x5d, 0x20, 0x7d}

	md := NewMsgDepack(bytes.NewReader(in))

	fmt.Printf("Here!\n")
	var element int
	for (element != Error) && (element != EOF) {
		element = md.Next()
		md.Data()
		switch element {
		case String:
			fmt.Printf("String, %s\n", *md.PayloadString)
		case Int64:
			fmt.Printf("Int64, %d\n", *md.PayloadInt64)
		case Map:
			fmt.Printf("Map, %d\n", *md.PayloadInt64)
		case Error:
			fmt.Printf("More, Error\n")
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	in := []byte{0x85, 0xa8, 0x6d, 0x73, 0x67, 0x5f, 0x74, 0x79,
		0x70, 0x65, 0x03, 0xb0, 0x74, 0x72, 0x61, 0x6e,
		0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x75, 0x75,
		0x69, 0x64, 0xd9, 0x24, 0x39, 0x34, 0x34, 0x37,
		0x32, 0x34, 0x31, 0x63, 0x2d, 0x35, 0x32, 0x33,
		0x38, 0x2d, 0x34, 0x63, 0x62, 0x39, 0x2d, 0x39,
		0x62, 0x61, 0x61, 0x2d, 0x37, 0x30, 0x37, 0x36,
		0x65, 0x33, 0x32, 0x33, 0x32, 0x38, 0x39, 0x39,
		0xa6, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0xd9,
		0x26, 0x64, 0x6e, 0x73, 0x3a, 0x77, 0x65, 0x62,
		0x70, 0x61, 0x2e, 0x63, 0x6f, 0x6d, 0x63, 0x61,
		0x73, 0x74, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76,
		0x32, 0x2d, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65,
		0x2d, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0xa4,
		0x64, 0x65, 0x73, 0x74, 0xb2, 0x73, 0x65, 0x72,
		0x69, 0x61, 0x6c, 0x3a, 0x31, 0x32, 0x33, 0x34,
		0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0xa7,
		0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0xc4,
		0x45, 0x7b, 0x20, 0x22, 0x6e, 0x61, 0x6d, 0x65,
		0x73, 0x22, 0x3a, 0x20, 0x5b, 0x20, 0x22, 0x44,
		0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x58, 0x5f,
		0x43, 0x49, 0x53, 0x43, 0x4f, 0x5f, 0x43, 0x4f,
		0x4d, 0x5f, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69,
		0x74, 0x79, 0x2e, 0x46, 0x69, 0x72, 0x65, 0x77,
		0x61, 0x6c, 0x6c, 0x2e, 0x46, 0x69, 0x72, 0x65,
		0x77, 0x61, 0x6c, 0x6c, 0x4c, 0x65, 0x76, 0x65,
		0x6c, 0x22, 0x20, 0x5d, 0x20, 0x7d}

	for i := 0; i < b.N; i++ {
		md := NewMsgDepack(bytes.NewReader(in))

		b.StartTimer()
		var element int
		for (element != Error) && (element != EOF) {
			element = md.Next()
		}
		b.StopTimer()
	}
}
