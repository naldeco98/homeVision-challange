package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"homeVision/internal/parser"
)

func main() {
	maxMetaSize := flag.Int64("max-meta-size", 10*1024*1024, "Maximum allowed size for metadata in bytes")
	flag.Parse()

	inputFile := "sample.env"
	outputDir := "output"

	// Ensure output directory exists
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

	p := parser.New(f, *maxMetaSize)

	for {
		chunk, err := p.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				fmt.Println("Warning: Reached end of file with trailing bytes.")
			} else {
				fmt.Fprintf(os.Stderr, "Error parsing chunk: %v\n", err)
			}
			break
		}

		// Process Content
		if chunk.Tag == "**%%DOCU" {
			filename, ok := chunk.Metadata["FILENAME"]
			if !ok {
				fmt.Println("Warning: DOCU chunk without FILENAME")
				// Skip content
				if err := skipContent(f, chunk.ContentLen); err != nil {
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
			_, err = io.CopyN(outFile, f, int64(chunk.ContentLen))
			outFile.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing content to file: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Extracted: %s (%d bytes)\n", filename, chunk.ContentLen)
		} else {
			// Skip content for other chunks
			if err := skipContent(f, chunk.ContentLen); err != nil {
				fmt.Fprintf(os.Stderr, "Error skipping content: %v\n", err)
				os.Exit(1)
			}
		}
	}
	fmt.Println("Extraction complete.")
}

func skipContent(seeker io.Seeker, length uint32) error {
	_, err := seeker.Seek(int64(length), io.SeekCurrent)
	return err
}
