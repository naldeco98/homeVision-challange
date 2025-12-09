package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	maxMetaSize := flag.Int64("max-meta-size", 10*1024*1024, "Maximum allowed size for metadata in bytes")
	flag.Parse()

	f, err := os.Open("sample.env")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "Error reading tag: %v\n", err)
			os.Exit(1)
		}

		// Read Meta Length
		var metaLen uint32
		err = binary.Read(f, binary.LittleEndian, &metaLen)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading meta length: %v\n", err)
			os.Exit(1)
		}

		if int64(metaLen) > *maxMetaSize {
			fmt.Fprintf(os.Stderr, "Error: metadata length %d exceeds maximum allowed size %d\n", metaLen, *maxMetaSize)
			os.Exit(1)
		}

		fmt.Printf("Offset: 0x%x, Tag: %s, MetaLen: %d\n", offset, string(tag), metaLen)

		// Read Metadata
		meta := make([]byte, metaLen)
		_, err = io.ReadFull(f, meta)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading metadata: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  Metadata: %q\n", string(meta))

		// Read Content Length
		var contentLen uint32
		err = binary.Read(f, binary.LittleEndian, &contentLen)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading content length: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("  ContentLen: %d\n", contentLen)

		// Skip Content
		_, err = f.Seek(int64(contentLen), io.SeekCurrent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error skipping content: %v\n", err)
			os.Exit(1)
		}
	}
}
