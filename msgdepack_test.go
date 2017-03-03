package msgdepack

import (
	"bytes"
	"fmt"
	"testing"
)

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