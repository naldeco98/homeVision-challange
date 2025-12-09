package parser

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestParser_Next_ValidChunk(t *testing.T) {
	tagStr := "**%%DOCU"
	var tag [8]byte
	copy(tag[:], tagStr)
	metaData := "FILENAME/test.txt\nSIZE/100"
	metaLen := uint32(len(metaData))
	contentLen := uint32(10)

	buf := new(bytes.Buffer)
	// Write Header
	binary.Write(buf, binary.LittleEndian, tag)
	binary.Write(buf, binary.LittleEndian, metaLen)
	// Write Metadata
	buf.WriteString(metaData)
	// Write ContentLen
	binary.Write(buf, binary.LittleEndian, contentLen)
	// Write Content (simulating existing content)
	buf.Write(bytes.Repeat([]byte("A"), int(contentLen)))

	p := New(buf, 1024)
	chunk, err := p.Next()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if chunk.Tag != tagStr {
		t.Errorf("Expected tag %q, got %q", tagStr, chunk.Tag)
	}
	if chunk.ContentLen != contentLen {
		t.Errorf("Expected ContentLen %d, got %d", contentLen, chunk.ContentLen)
	}
	if chunk.Metadata["FILENAME"] != "test.txt" {
		t.Errorf("Expected metadata FILENAME='test.txt', got %q", chunk.Metadata["FILENAME"])
	}
}

func TestParser_Next_MetaLenTooLarge(t *testing.T) {
	tagStr := "**%%DOCU"
	var tag [8]byte
	copy(tag[:], tagStr)
	metaLen := uint32(2000) // Larger than maxMetaSize 1024

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, tag)
	binary.Write(buf, binary.LittleEndian, metaLen)

	p := New(buf, 1024)
	_, err := p.Next()
	if err == nil {
		t.Fatal("Expected error for large MetaLen, got nil")
	}
	if !errors.Is(err, ErrMetaTooLarge) {
		t.Errorf("Expected error ErrMetaTooLarge, got %v", err)
	}
}

func TestParser_Next_DetailedErrors(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() io.Reader
		expectErr   bool
		errContains string
	}{
		{
			name: "Invalid Tag",
			setup: func() io.Reader {
				buf := new(bytes.Buffer)
				buf.Write([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}) // Non-printable
				binary.Write(buf, binary.LittleEndian, uint32(10))
				return buf
			},
			expectErr:   true,
			errContains: "invalid tag",
		},
		{
			name: "Incomplete Metadata",
			setup: func() io.Reader {
				tagStr := "TAG12345"
				var tag [8]byte
				copy(tag[:], tagStr)
				metaLen := uint32(10)
				buf := new(bytes.Buffer)
				binary.Write(buf, binary.LittleEndian, tag)
				binary.Write(buf, binary.LittleEndian, metaLen)
				buf.WriteString("123") // Only 3 bytes
				return buf
			},
			expectErr:   true,
			errContains: "unexpected EOF",
		},
		{
			name: "Incomplete ContentLen",
			setup: func() io.Reader {
				tagStr := "TAG12345"
				var tag [8]byte
				copy(tag[:], tagStr)
				metaData := "KEY/VAL"
				metaLen := uint32(len(metaData))
				buf := new(bytes.Buffer)
				binary.Write(buf, binary.LittleEndian, tag)
				binary.Write(buf, binary.LittleEndian, metaLen)
				buf.WriteString(metaData)
				buf.Write([]byte{0x01}) // Only 1 byte for uint32
				return buf
			},
			expectErr:   true,
			errContains: "unexpected EOF",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := New(tc.setup(), 1024)
			_, err := p.Next()
			if !tc.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tc.expectErr {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tc.errContains != "" {
					if !strings.Contains(err.Error(), tc.errContains) {
						t.Errorf("Expected error %q to contain %q", err.Error(), tc.errContains)
					}
				}
			}
		})
	}
}
