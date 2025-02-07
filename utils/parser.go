package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

func ParseError(err error) {
	fmt.Println("[=] Error:", err.Error())
}

type StringOptions struct {
	MinCharacters    int
	PrintNormal      bool
	PrintAsciiOnly   bool
	PrintUnicodeOnly bool
}

const (
	TYPE_UNDETERMINED = iota
	TYPE_ASCII
	TYPE_UNICODE
)

const (
	EXTRACT_RAW = iota
	EXTRACT_ASM
)

const (
	MAX_STRING_SIZE = 4096
	BLOCK_SIZE      = 0x50000
)

type PrintBuffer struct {
	buffer *bytes.Buffer
}

func NewPrintBuffer() *PrintBuffer {
	return &PrintBuffer{
		buffer: bytes.NewBuffer(nil),
	}
}

func (pb *PrintBuffer) AddString(s string) {
	pb.buffer.WriteString(s)
}

func (pb *PrintBuffer) Flush() {
	fmt.Print(pb.buffer.String())
	pb.buffer.Reset()
}

type StringParser struct {
	options StringOptions
	printer *PrintBuffer
	isAscii [256]bool
}

func NewStringParser(options StringOptions) *StringParser {
	sp := &StringParser{
		options: options,
		printer: NewPrintBuffer(),
	}
	for i := 0; i < 256; i++ {
		sp.isAscii[i] = i >= 0x20 && i <= 0x7E
	}
	return sp
}

func (sp *StringParser) extractImmediate(immediate []byte, strType int) (int, int, []byte) {
	i := 0
	output := make([]byte, 0, len(immediate))

	switch strType {
	case TYPE_ASCII:
		for i < len(immediate) && sp.isAscii[immediate[i]] {
			output = append(output, immediate[i])
			i++
		}
		return i, strType, output

	case TYPE_UNICODE:
		for i+1 < len(immediate) && sp.isAscii[immediate[i]] && immediate[i+1] == 0 {
			output = append(output, immediate[i])
			i += 2
		}
		return i / 2, strType, output

	case TYPE_UNDETERMINED:
		if len(immediate) == 0 || !sp.isAscii[immediate[0]] {
			return 0, strType, output
		}
		if len(immediate) > 1 && immediate[1] == 0 {
			return sp.extractImmediate(immediate, TYPE_UNICODE)
		}
		return sp.extractImmediate(immediate, TYPE_ASCII)

	default:
		return 0, strType, output
	}
}

func (sp *StringParser) extractString(buffer []byte, offset int) (int, int, int, []byte) {
	if offset+3 >= len(buffer) {
		return 0, EXTRACT_RAW, TYPE_UNDETERMINED, nil
	}

	value := binary.LittleEndian.Uint16(buffer[offset:])
	output := make([]byte, 0)
	extractType := EXTRACT_RAW
	strType := TYPE_UNDETERMINED
	processed := 0

	switch value {
	case 0x45C6:
		instSize := 4
		immSize := 1
		immOffset := 3

		for offset+processed+instSize <= len(buffer) {
			if buffer[offset+processed] != 0xC6 || buffer[offset+processed+1] != 0x45 {
				break
			}

			imm := buffer[offset+processed+immOffset : offset+processed+instSize]
			p, t, o := sp.extractImmediate(imm, strType)
			output = append(output, o...)
			strType = t

			if (strType == TYPE_UNICODE && p < (immSize+1)/2) || (strType == TYPE_ASCII && p < immSize) {
				break
			}

			processed += instSize
		}
		extractType = EXTRACT_ASM

	case 0x85C6:
		instSize := 8
		immSize := 1
		immOffset := 7

		for offset+processed+instSize <= len(buffer) {
			if buffer[offset+processed] != 0xC6 || buffer[offset+processed+1] != 0x85 {
				break
			}

			imm := buffer[offset+processed+immOffset : offset+processed+instSize]
			p, t, o := sp.extractImmediate(imm, strType)
			output = append(output, o...)
			strType = t

			if (strType == TYPE_UNICODE && p < (immSize+1)/2) || (strType == TYPE_ASCII && p < immSize) {
				break
			}

			processed += instSize
		}
		extractType = EXTRACT_ASM

	case 0x45C7:
		instSize := 7
		immSize := 4
		immOffset := 3

		for offset+processed+instSize <= len(buffer) {
			if buffer[offset+processed] != 0xC7 || buffer[offset+processed+1] != 0x45 {
				break
			}

			imm := buffer[offset+processed+immOffset : offset+processed+instSize]
			p, t, o := sp.extractImmediate(imm, strType)
			output = append(output, o...)
			strType = t

			if (strType == TYPE_UNICODE && p < (immSize+1)/2) || (strType == TYPE_ASCII && p < immSize) {
				break
			}

			processed += instSize
		}
		extractType = EXTRACT_ASM

	case 0x85C7:
		instSize := 10
		immSize := 4
		immOffset := 6

		for offset+processed+instSize <= len(buffer) {
			if buffer[offset+processed] != 0xC7 || buffer[offset+processed+1] != 0x85 {
				break
			}

			imm := buffer[offset+processed+immOffset : offset+processed+instSize]
			p, t, o := sp.extractImmediate(imm, strType)
			output = append(output, o...)
			strType = t

			if (strType == TYPE_UNICODE && p < (immSize+1)/2) || (strType == TYPE_ASCII && p < immSize) {
				break
			}

			processed += instSize
		}
		extractType = EXTRACT_ASM

	default:
		if sp.isAscii[buffer[offset]] {
			if buffer[offset+1] == 0 {
				i := 0
				for offset+i+1 < len(buffer) && sp.isAscii[buffer[offset+i]] && buffer[offset+i+1] == 0 {
					output = append(output, buffer[offset+i])
					i += 2
				}
				return i, EXTRACT_RAW, TYPE_UNICODE, output
			} else {
				i := 0
				for offset+i < len(buffer) && sp.isAscii[buffer[offset+i]] {
					output = append(output, buffer[offset+i])
					i++
				}
				return i, EXTRACT_RAW, TYPE_ASCII, output
			}
		}
	}

	return processed, extractType, strType, output
}

func (sp *StringParser) processContents(buffer []byte) bool {
	offset := 0

	for offset < len(buffer) {
		processed, extractType, strType, str := sp.extractString(buffer, offset)
		if processed == 0 {
			offset++
			continue
		}

		if len(str) >= sp.options.MinCharacters {
			print := false
			if sp.options.PrintNormal && extractType == EXTRACT_RAW {
				print = true
			}
			if sp.options.PrintAsciiOnly && strType != TYPE_ASCII {
				print = false
			}
			if sp.options.PrintUnicodeOnly && strType != TYPE_UNICODE {
				print = false
			}

			if print {
				s := strings.ReplaceAll(string(str), "\n", "\\n")
				s = strings.ReplaceAll(s, "\r", "\\r")
				sp.printer.AddString(s + "\n")
			}
		}

		offset += processed
	}

	sp.printer.Flush()
	return true
}

func (sp *StringParser) ParseBlock(buffer []byte) bool {
	if len(buffer) == 0 {
		return false
	}
	return sp.processContents(buffer)
}
