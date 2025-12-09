package parser

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

var (
	// ErrMetaTooLarge is returned when the metadata length exceeds the maximum allowed size.
	ErrMetaTooLarge = errors.New("metadata exceeds maximum size")
	// ErrInvalidTag is returned when a chunk tag contains non-printable characters.
	ErrInvalidTag = errors.New("invalid tag found")
)

// ChunkHeader represents the fixed-size header of a chunk
type ChunkHeader struct {
	Tag     [8]byte
	MetaLen uint32
}

// Chunk contains the parsed information of a chunk, excluding the actual content data.
type Chunk struct {
	Tag        string
	MetaLen    uint32
	Metadata   map[string]string
	ContentLen uint32
}

// Parser handles the parsing of the custom file format.
type Parser struct {
	r           io.Reader
	maxMetaSize int64
}

// New creates a new Parser.
func New(r io.Reader, maxMetaSize int64) *Parser {
	return &Parser{
		r:           r,
		maxMetaSize: maxMetaSize,
	}
}

// Next parses the next chunk header and metadata.
// It returns the Chunk info and an error.
// The caller is responsible for reading or skipping chunks (ContentLen bytes)
// from the underlying reader before calling Next again.
func (p *Parser) Next() (*Chunk, error) {
	var header ChunkHeader
	// Read Tag and MetaLen
	err := binary.Read(p.r, binary.LittleEndian, &header)
	if err == io.EOF {
		return nil, io.EOF
	}
	if err != nil {
		return nil, fmt.Errorf("reading chunk header: %w", err)
	}

	// Validate MetaLen against maxMetaSize to prevent large allocations
	if int64(header.MetaLen) > p.maxMetaSize {
		return nil, ErrMetaTooLarge
	}

	tag := string(header.Tag[:])
	if !isValidTag(tag) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidTag, tag)
	}

	// Read Metadata
	// We allocate the buffer here. Since we checked MetaLen <= maxMetaSize, this is safe from DoS.
	metaBytes := make([]byte, header.MetaLen)
	_, err = io.ReadFull(p.r, metaBytes)
	if err != nil {
		if err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, fmt.Errorf("reading metadata: %w", err)
	}

	metadata := parseMetadata(metaBytes)

	// Read Content Length
	var contentLen uint32
	err = binary.Read(p.r, binary.LittleEndian, &contentLen)
	if err != nil {
		if err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, fmt.Errorf("reading content length: %w", err)
	}

	return &Chunk{
		Tag:        tag,
		MetaLen:    header.MetaLen,
		Metadata:   metadata,
		ContentLen: contentLen,
	}, nil
}

func isValidTag(tag string) bool {
	for _, r := range tag {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func parseMetadata(data []byte) map[string]string {
	m := make(map[string]string)
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
