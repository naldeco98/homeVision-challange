package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"homeVision/internal/parser"
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

	p := parser.New(f, *maxMetaSize)

	for {
		offset, _ := f.Seek(0, io.SeekCurrent)

		chunk, err := p.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				fmt.Printf("Trailing bytes at offset 0x%x\n", offset)
				break
			}
			fmt.Fprintf(os.Stderr, "Error parsing chunk at offset 0x%x: %v\n", offset, err)
			os.Exit(1)
		}

		fmt.Printf("Offset: 0x%x, Tag: %s, MetaLen: %d\n", offset, chunk.Tag, chunk.MetaLen)
		// Note: The original analyzer printed MetaLen from header.
		// Our parser returns Metadata map, not raw length (although we read it).
		// We can't easily get exact raw MetaLen from chunk anymore unless we add it to Chunk struct.
		// However, for analysis, seeing Tag and ContentLen is useful.
		// Let's modify Chunk struct to include MetaLen if we really want it, or just ignore for now.
		// The original analyzer printed: `Offset: 0x%x, Tag: %s, MetaLen: %d`
		// Then `Metadata: %q`.

		fmt.Printf("  Metadata: %v\n", chunk.Metadata)
		fmt.Printf("  ContentLen: %d\n", chunk.ContentLen)

		// Skip Content
		_, err = f.Seek(int64(chunk.ContentLen), io.SeekCurrent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error skipping content: %v\n", err)
			os.Exit(1)
		}
	}
}
