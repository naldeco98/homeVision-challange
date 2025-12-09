# HomeVision Challenge Solution

This repository contains the solution for the HomeVision Take-home Challenge.

## Problem
Reverse engineer a custom `.env` file format and extract its contents.

## Solution
The solution is implemented in Go. The main parser is located in `cmd/main.go`. It reads the `.env` file, parses the custom chunk-based format, and extracts the contained files to an `output` directory.

### File Format Analysis
The `.env` file consists of a sequence of chunks. Each chunk has the following structure:

1.  **Tag** (8 bytes): ASCII string identifying the chunk type (e.g., `DC%%STAM`, `**%%KEYB`, `**%%DOCU`).
2.  **Metadata Length** (4 bytes): Little-endian uint32 specifying the length of the metadata block.
3.  **Metadata** (Variable length): Text data containing key-value pairs separated by newlines. Keys and values are separated by `/`.
4.  **Content Length** (4 bytes): Little-endian uint32 specifying the length of the content block.
5.  **Content** (Variable length): Binary data of the file or section content.

The `**%%DOCU` chunks contain the actual files. The metadata in these chunks includes the `FILENAME` which is used for extraction.

### Usage
To run the parser:

```bash
go run cmd/main.go
```

This will create an `output` directory and extract the files there.

### Analyzer Tool
A separate analyzer tool is available to inspect the raw chunks of the `.env` file without extracting content. This is useful for debugging or understanding the file structure.

To run the analyzer:

```bash
go run internal/analyzer/main.go
```

This will create an `output` directory and extract the files there.

## Bonus: Production Considerations
If this were a production release, I would make the following improvements:

1.  **Streaming & Memory Efficiency**: The current implementation reads the entire metadata block into memory. For extremely large metadata or files, I would implement a streaming reader to handle data in smaller buffers.
2.  **Validation**: The metadata includes `SHA1` checksums. A production parser should calculate the SHA1 of the extracted content and verify it against the metadata to ensure data integrity.
3.  **Error Handling**: Replace `panic` with proper error handling and logging. This would allow the program to gracefully handle malformed chunks or file I/O errors without crashing.
4.  **Configuration**: Add command-line flags (using `flag` package) to specify input file path, output directory, and verbosity.
5.  **Testing**: Add unit tests for the parsing logic, specifically for the metadata parsing and chunk reading. Fuzz testing could also be used to ensure robustness against malformed inputs.
6.  **Concurrency**: While the file format is sequential, file writing could be decoupled from reading using a worker pool pattern to speed up extraction on systems with fast I/O.
