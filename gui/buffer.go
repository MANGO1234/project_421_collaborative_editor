package gui

import "bytes"

type ByteSequence interface {
	NextWord() []byte
}

type StringSequence struct {
	str string
	pos int
}

func NewStringSequence(str string) *StringSequence {
	return &StringSequence{str, 0}
}

func (seq *StringSequence) HasNext() bool {
	return seq.pos < len(seq.str)
}

func (seq *StringSequence) NextWord() []byte {
	if seq.pos >= len(seq.str) {
		return nil
	}

	ch := seq.str[seq.pos]
	if ch == ' ' {
		seq.pos++
		return []byte(" ")
	}
	if ch == '\t' {
		seq.pos++
		return []byte("\t")
	}
	if ch == '\n' {
		seq.pos++
		return []byte("\n")
	}

	start := seq.pos
	end := seq.pos + 1
	for end < len(seq.str) {
		ch := seq.str[end]
		if ch == ' ' || ch == '\t' || ch == '\n' {
			break
		}
		end++
	}
	seq.pos = end
	return []byte(seq.str[start:end])
}

type Line struct {
	prev  *Line
	next  *Line
	bytes []byte
}

type Buffer struct {
	cursorx   int
	cursory   int
	bufferpos int
	width     int
	lines     *Line
}

func StringToBuffer(str string, width int) *Buffer {
	buf := &Buffer{cursorx: 0, cursory: 0, bufferpos: 0, width: width}
	buf.lines = SeqToLines(NewStringSequence(str), width)
	return buf
}

func WordLength(word []byte) int {
	if word[0] == '\t' {
		return 4
	} else {
		return len(word)
	}
}

func SeqToLines(seq ByteSequence, width int) *Line {
	line := &Line{bytes: make([]byte, 0, width+1)}
	startLine := line
	i := 0
	for {
		word := seq.NextWord()
		if word == nil {
			break
		}

		if word[0] == '\n' {
			line.bytes = append(line.bytes, '\n')
			lastLine := line
			line = &Line{bytes: make([]byte, 0, width), prev: lastLine}
			lastLine.next = line
			i = 0
			continue
		}

		wordLength := WordLength(word)
		if i+wordLength < width {
			line.bytes = append(line.bytes, word...)
			i += wordLength
			continue
		}

		if wordLength > width {
			line.bytes = append(line.bytes, word[:width-i]...)
			k := width - i
			for k < wordLength {
				lastLine := line
				line = &Line{bytes: make([]byte, 0, width), prev: lastLine}
				lastLine.next = line
				line.bytes = append(line.bytes, word[k:k+width]...)
				k += width
			}
			continue
		} else {
			lastLine := line
			line = &Line{bytes: make([]byte, 0, width), prev: lastLine}
			lastLine.next = line
			line.bytes = append(line.bytes, word...)
			i = wordLength
			continue
		}
	}
	return startLine
}

func (buf *Buffer) Lines() *Line {
	return buf.lines
}

func (buf *Buffer) ToString() string {
	builder := bytes.Buffer{}
	lines := buf.lines
	for lines != nil {
		builder.Write(lines.bytes)
		lines = lines.next
	}
	return builder.String()
}

func NumberOfLines(lines *Line) int {
	i := 0
	for lines != nil {
		i++
		lines = lines.next
	}
	return i
}
