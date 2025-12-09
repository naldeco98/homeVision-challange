package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func main() {
	f, err := os.Open("sample.env")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for {
		offset, _ := f.Seek(0, io.SeekCurrent)

		// Read Tag
		tag := make([]byte, 8)
		_, err := io.ReadFull(f, tag)
		if err == io.EOF {
			break
		}
		if err != nil {
			// Check if we are at the end with few bytes left
			if err == io.ErrUnexpectedEOF {
				fmt.Printf("Trailing bytes at offset 0x%x\n", offset)
				break
			}
			panic(err)
		}

		// Read Meta Length
		var metaLen uint32
		err = binary.Read(f, binary.LittleEndian, &metaLen)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Offset: 0x%x, Tag: %s, MetaLen: %d\n", offset, string(tag), metaLen)

		// Read Metadata
		meta := make([]byte, metaLen)
		_, err = io.ReadFull(f, meta)
		if err != nil {
			panic(err)
		}
		fmt.Printf("  Metadata: %q\n", string(meta))

		// Read Content Length
		var contentLen uint32
		err = binary.Read(f, binary.LittleEndian, &contentLen)
		if err != nil {
			panic(err)
		}

		fmt.Printf("  ContentLen: %d\n", contentLen)

		// Skip Content
		_, err = f.Seek(int64(contentLen), io.SeekCurrent)
		if err != nil {
			panic(err)
		}
	}
}
