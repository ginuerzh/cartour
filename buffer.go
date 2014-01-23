// buffer
package main

import (
	"strings"
)

type ContentBuffer struct {
	valid bool
	lines []string
}

func NewContentBuffer() *ContentBuffer {
	return &ContentBuffer{false, make([]string, 1)}
}

func (buffer *ContentBuffer) Lines() int {
	return len(buffer.lines)
}

func (buffer *ContentBuffer) currentLine() string {
	return buffer.lines[buffer.Lines()-1]
}

func (buffer *ContentBuffer) Newline() {
	buffer.lines[buffer.Lines()-1] = strings.TrimSpace(buffer.currentLine())
	if len(buffer.currentLine()) == 0 {
		return
	}
	buffer.lines = append(buffer.lines, "")
}

func (buffer *ContentBuffer) Append(s string) {
	if len(s) == 0 {
		return
	}
	cur := buffer.currentLine()
	buffer.lines[buffer.Lines()-1] = cur + s
}

func (buffer *ContentBuffer) Valid(b bool) {
	buffer.valid = b
}

func (buffer *ContentBuffer) IsValid() bool {
	return buffer.valid
}

func (buffer *ContentBuffer) Content() []string {
	return buffer.lines
}
