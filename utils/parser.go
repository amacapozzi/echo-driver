package utils

import (
	"fmt"
)

func ParseError(err error) {
	fmt.Println("[=] Error:", err.Error())
}

type StringOptions struct {
	MinChars            int
	PrintNotInteresting bool
	PrintJSON           bool
	PrintFilepath       bool
	PrintFilename       bool
	PrintStringType     bool
	PrintSpan           bool
	EscapeNewLines      bool
	OffsetStart         int64
	OffsetEnd           int64
}

type StringParser struct {
	Options StringOptions
	Printer *PrintBuffer
}

type PrintBuffer struct {
	Buffer       []byte
	SpaceUsed    int
	IsStart      bool
	AddJSONClose bool
}

type StringInfo struct {
	String        string
	Type          string
	Span          [2]uintptr
	IsInteresting bool
}

func NewStringParser(options StringOptions) *StringParser {
	return &StringParser{
		Options: options,
		Printer: &PrintBuffer{
			Buffer: make([]byte, 0x100000),
		},
	}
}

func (sp *StringParser) ParseBlock(buffer []byte, nameShort, nameLong string, baseAddress uintptr) bool {
	if len(buffer) > 0 {
		rVect := sp.ExtractAllStrings(buffer)

		for _, r := range rVect {
			sp.Printer.AddString(r.String + "\n")
		}
	}
	return false
}
func (sp *StringParser) ExtractAllStrings(buffer []byte) []StringInfo {
	var results []StringInfo
	minChars := sp.Options.MinChars

	start := 0
	for i := 0; i < len(buffer); i++ {
		if buffer[i] >= 32 && buffer[i] <= 126 {
			continue
		}

		if i-start >= minChars {
			str := string(buffer[start:i])
			results = append(results, StringInfo{
				String:        str,
				Type:          "ASCII",
				Span:          [2]uintptr{uintptr(start), uintptr(i)},
				IsInteresting: true,
			})
		}

		start = i + 1
	}

	if len(buffer)-start >= minChars {
		str := string(buffer[start:])
		results = append(results, StringInfo{
			String:        str,
			Type:          "ASCII",
			Span:          [2]uintptr{uintptr(start), uintptr(len(buffer))},
			IsInteresting: true,
		})
	}

	return results
}
func (pb *PrintBuffer) AddString(s string) {
	if pb.SpaceUsed+len(s) >= len(pb.Buffer) {
		pb.Digest()
	}
	if pb.SpaceUsed+len(s) < len(pb.Buffer) {
		copy(pb.Buffer[pb.SpaceUsed:], s)
		pb.SpaceUsed += len(s)
	} else {
		fmt.Print(s)
	}
}

func (pb *PrintBuffer) AddJSONString(jsonStr string) {
	if pb.IsStart {
		pb.AddString("[" + jsonStr)
		pb.IsStart = false
		pb.AddJSONClose = true
	} else {
		pb.AddString("," + jsonStr)
	}
}

func (pb *PrintBuffer) Digest() {
	if pb.SpaceUsed > 0 {
		fmt.Print(string(pb.Buffer[:pb.SpaceUsed]))
		pb.SpaceUsed = 0
	}
	if pb.AddJSONClose {
		pb.AddString("]")
		pb.AddJSONClose = false
	}
}
