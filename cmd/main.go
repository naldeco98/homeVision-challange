package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// ChunkHeader represents the fixed-size header of a chunk
type ChunkHeader struct {
	Tag     [8]byte
	MetaLen uint32
}

// Metadata represents the parsed metadata of a chunk
type Metadata map[string]string

func main() {
	maxMetaSize := flag.Int64("max-meta-size", 10*1024*1024, "Maximum allowed size for metadata in bytes")
	flag.Parse()

	inputFile := "sample.env"
	outputDir := "output"

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	f, err := os.Open(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	for {
		// Read Tag and MetaLen
		var header ChunkHeader
		err := binary.Read(f, binary.LittleEndian, &header)
		if err == io.EOF {
			break
		}
		if err != nil {
			// Check for trailing bytes or unexpected EOF
			if err == io.ErrUnexpectedEOF {
				fmt.Println("Reached end of file with trailing bytes.")
				break
			}
			fmt.Fprintf(os.Stderr, "Error reading header: %v\n", err)
			os.Exit(1)
		}

		if int64(header.MetaLen) > *maxMetaSize {
			fmt.Fprintf(os.Stderr, "Error: metadata length %d exceeds maximum allowed size %d\n", header.MetaLen, *maxMetaSize)
			os.Exit(1)
		}

		tag := string(header.Tag[:])
		if !isValidTag(tag) {
			fmt.Fprintf(os.Stderr, "Error: invalid tag found: %q\n", tag)
			os.Exit(1)
		}

		// Read Metadata
		metaBytes := make([]byte, header.MetaLen)
		_, err = io.ReadFull(f, metaBytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading metadata: %v\n", err)
			os.Exit(1)
		}

		metadata := parseMetadata(metaBytes)

		// Read Content Length
		var contentLen uint32
		err = binary.Read(f, binary.LittleEndian, &contentLen)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading content length: %v\n", err)
			os.Exit(1)
		}

		// Process Content
		if tag == "**%%DOCU" {
			filename, ok := metadata["FILENAME"]
			if !ok {
				fmt.Println("Warning: DOCU chunk without FILENAME")
				// Skip content
				_, err = f.Seek(int64(contentLen), io.SeekCurrent)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error skipping content: %v\n", err)
					os.Exit(1)
				}
				continue
			}

			outputPath := filepath.Join(outputDir, filename)
			outFile, err := os.Create(outputPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
				os.Exit(1)
			}

			// Copy content to file
			_, err = io.CopyN(outFile, f, int64(contentLen))
			outFile.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing content to file: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Extracted: %s (%d bytes)\n", filename, contentLen)
		} else {
			// Skip content for other chunks
			_, err = f.Seek(int64(contentLen), io.SeekCurrent)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error skipping content: %v\n", err)
				os.Exit(1)
			}
		}
	}
	fmt.Println("Extraction complete.")
}

func isValidTag(tag string) bool {
	for _, r := range tag {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func parseMetadata(data []byte) Metadata {
	m := make(Metadata)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "/", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}
