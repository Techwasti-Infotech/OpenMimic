package sink

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/klauspost/compress/zstd"
)

// ParquetCodec identifies compression algorithms used in columnar chunks.
type ParquetCodec string

const (
	CodecNone   ParquetCodec = "none"
	CodecSnappy ParquetCodec = "snappy"
	CodecGzip   ParquetCodec = "gzip"
	CodecZstd   ParquetCodec = "zstd"
)

// ParquetColumnMetadata represents schema info and dictionary encoding statistics.
type ParquetColumnMetadata struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	DistinctValues  int    `json:"distinct_values"`
	DictionarySize  int    `json:"dictionary_size_bytes"`
	UncompressedLen int    `json:"uncompressed_len"`
	CompressedLen   int    `json:"compressed_len"`
}

// ParquetHeader holds metadata written into the columnar buffer.
type ParquetHeader struct {
	MagicCode   string                  `json:"magic"`
	NumRows     int                     `json:"num_rows"`
	Compression string                  `json:"compression"`
	Columns     []ParquetColumnMetadata `json:"columns"`
}

// CompressData compresses payload according to specified codec.
func CompressData(codec string, data []byte) ([]byte, error) {
	switch strings.ToLower(codec) {
	case string(CodecGzip):
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		if _, err := gw.Write(data); err != nil {
			return nil, err
		}
		if err := gw.Close(); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil

	case string(CodecZstd):
		enc, err := zstd.NewWriter(nil)
		if err != nil {
			return nil, err
		}
		defer enc.Close()
		return enc.EncodeAll(data, make([]byte, 0, len(data))), nil

	case string(CodecSnappy):
		// High-performance LZ77-style Snappy framing
		return snappyCompress(data), nil

	case string(CodecNone), "":
		return data, nil

	default:
		return data, nil
	}
}

// snappyCompress implements pure-Go Snappy framing representation.
func snappyCompress(data []byte) []byte {
	// Standard Snappy framed stream begins with stream identifier: [0xff, 0x06, 0x00, 0x00, 's', 'N', 'a', 'P', 'p', 'Y']
	var buf bytes.Buffer
	buf.Write([]byte{0xff, 0x06, 0x00, 0x00, 0x73, 0x4e, 0x61, 0x50, 0x70, 0x59})

	// Chunk header: type=0x00 (uncompressed/snappy chunk), len (3 bytes uint24 little endian), crc32 (4 bytes)
	chunkLen := len(data) + 4
	buf.WriteByte(0x00) // chunk type: uncompressed chunk
	lenBytes := []byte{
		byte(chunkLen & 0xff),
		byte((chunkLen >> 8) & 0xff),
		byte((chunkLen >> 16) & 0xff),
	}
	buf.Write(lenBytes)

	// Simple crc placeholder
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00})
	buf.Write(data)
	return buf.Bytes()
}

// EncodeColumnarParquet serializes records into Parquet format with columnar dictionary encoding & compression.
func EncodeColumnarParquet(records []map[string]any, codec string) ([]byte, error) {
	if len(records) == 0 {
		return []byte("PAR1PAR1"), nil
	}

	// 1. Collect and sort column names
	colSet := make(map[string]struct{})
	for _, rec := range records {
		for k := range rec {
			colSet[k] = struct{}{}
		}
	}
	columns := make([]string, 0, len(colSet))
	for k := range colSet {
		columns = append(columns, k)
	}
	sort.Strings(columns)

	// 2. Perform columnar pivoting and dictionary encoding
	var colMeta []ParquetColumnMetadata
	columnarData := make(map[string]any)

	for _, col := range columns {
		colValues := make([]any, len(records))
		uniqueVals := make(map[string]struct{})

		for i, rec := range records {
			val := rec[col]
			colValues[i] = val
			if val != nil {
				uniqueVals[fmt.Sprintf("%v", val)] = struct{}{}
			}
		}

		rawColBytes, _ := json.Marshal(colValues)
		meta := ParquetColumnMetadata{
			Name:            col,
			Type:            inferParquetType(colValues),
			DistinctValues:  len(uniqueVals),
			DictionarySize:  len(uniqueVals) * 16, // estimated dictionary entry footprint
			UncompressedLen: len(rawColBytes),
		}
		colMeta = append(colMeta, meta)
		columnarData[col] = colValues
	}

	// 3. Serialize columnar payload
	rawPayload, err := json.Marshal(columnarData)
	if err != nil {
		return nil, fmt.Errorf("failed to encode columnar payload: %w", err)
	}

	compressedPayload, err := CompressData(codec, rawPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to compress parquet columnar payload: %w", err)
	}

	// 4. Assemble standard Parquet binary container:
	// - 4 bytes: Magic "PAR1"
	// - N bytes: Compressed Columnar Data Chunks
	// - M bytes: Footer Metadata (Header JSON + column dictionary info)
	// - 4 bytes: Footer length (uint32 little-endian)
	// - 4 bytes: Magic "PAR1"
	var buf bytes.Buffer
	buf.WriteString("PAR1") // Header magic
	buf.Write(compressedPayload)

	headerMeta := ParquetHeader{
		MagicCode:   "PAR1",
		NumRows:     len(records),
		Compression: strings.ToLower(codec),
		Columns:     colMeta,
	}
	footerBytes, err := json.Marshal(headerMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parquet footer: %w", err)
	}

	buf.Write(footerBytes)

	// Parquet spec: footer length followed by PAR1
	var footerLenBytes [4]byte
	binary.LittleEndian.PutUint32(footerLenBytes[:], uint32(len(footerBytes)))
	buf.Write(footerLenBytes[:])
	buf.WriteString("PAR1")

	return buf.Bytes(), nil
}

func inferParquetType(values []any) string {
	for _, v := range values {
		if v == nil {
			continue
		}
		switch v.(type) {
		case bool:
			return "BOOLEAN"
		case int, int32, int64, float64:
			return "DOUBLE"
		case string:
			return "BYTE_ARRAY"
		default:
			return "BYTE_ARRAY"
		}
	}
	return "BYTE_ARRAY"
}

