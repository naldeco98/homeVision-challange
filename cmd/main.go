package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ChunkHeader represents the fixed-size header of a chunk
type ChunkHeader struct {
	Tag     [8]byte
	MetaLen uint32
}

// Metadata represents the parsed metadata of a chunk
type Metadata map[string]string

func main() {
	inputFile := "sample.env"
	outputDir := "output"

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		panic(err)
	}

	f, err := os.Open(inputFile)
	if err != nil {
		panic(err)
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
			panic(err)
		}

		tag := string(header.Tag[:])
		// fmt.Printf("Processing chunk: %s\n", tag)

		// Read Metadata
		metaBytes := make([]byte, header.MetaLen)
		_, err = io.ReadFull(f, metaBytes)
		if err != nil {
			panic(err)
		}

		metadata := parseMetadata(metaBytes)

		// Read Content Length
		var contentLen uint32
		err = binary.Read(f, binary.LittleEndian, &contentLen)
		if err != nil {
			panic(err)
		}

		// Process Content
		if tag == "**%%DOCU" {
			filename, ok := metadata["FILENAME"]
			if !ok {
				fmt.Println("Warning: DOCU chunk without FILENAME")
				// Skip content
				_, err = f.Seek(int64(contentLen), io.SeekCurrent)
				if err != nil {
					panic(err)
				}
				continue
			}

			outputPath := filepath.Join(outputDir, filename)
			outFile, err := os.Create(outputPath)
			if err != nil {
				panic(err)
			}

			// Copy content to file
			_, err = io.CopyN(outFile, f, int64(contentLen))
			outFile.Close()
			if err != nil {
				panic(err)
			}

			fmt.Printf("Extracted: %s (%d bytes)\n", filename, contentLen)
		} else {
			// Skip content for other chunks
			_, err = f.Seek(int64(contentLen), io.SeekCurrent)
			if err != nil {
				panic(err)
			}
		}
	}
	fmt.Println("Extraction complete.")
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
