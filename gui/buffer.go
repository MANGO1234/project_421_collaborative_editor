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

func wordLength(word []byte) int {
	if word[0] == '\t' {
		return 4
	} else {
		return len(word)
	}
}

func charLength(ch byte) int {
	if ch == '\t' {
		return 4
	} else if ch == '\n' {
		return 0
	} else {
		return 1
	}
}
func sliceLength(bytes []byte) int {
	s := 0
	for _, ch := range bytes {
		s += charLength(ch)
	}
	return s
}

type Line struct {
	prev  *Line
	next  *Line
	bytes []byte
}

type Buffer struct {
	width           int
	numberOfChars   int
	numberOfLines   int
	lines           *Line
	currentLine     *Line
	currentPosition int
	currentX        int
	currentY        int
}

func StringToBuffer(str string, width int) *Buffer {
	buf := &Buffer{width: width}
	buf.lines, buf.numberOfLines, buf.numberOfChars = SeqToLines(NewStringSequence(str), width)
	buf.currentLine = buf.lines
	buf.currentPosition = 0
	buf.currentX = 0
	buf.currentY = 0
	return buf
}

func SeqToLines(seq ByteSequence, width int) (*Line, int, int) {
	line := &Line{bytes: make([]byte, 0, width+1)}
	startLine := line
	i := 0
	numberOfLines := 1
	numberOfChars := 0
	for {
		word := seq.NextWord()
		numberOfChars += len(word)
		if word == nil {
			break
		}

		if word[0] == '\n' {
			line.bytes = append(line.bytes, '\n')
			lastLine := line
			line = &Line{bytes: make([]byte, 0, width), prev: lastLine}
			lastLine.next = line
			i = 0
			numberOfLines++
			continue
		}

		wordLength := wordLength(word)
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
				numberOfLines++
				k += width
			}
			i = width - (k - wordLength)
			continue
		} else {
			lastLine := line
			line = &Line{bytes: make([]byte, 0, width), prev: lastLine}
			lastLine.next = line
			line.bytes = append(line.bytes, word...)
			i = wordLength
			numberOfLines++
			continue
		}
	}
	// a sentinel \n at the end to remove edge cases from various methods
	line.bytes = append(line.bytes, '\n')
	return startLine, numberOfLines, numberOfChars
}

func (buf *Buffer) MoveLeft() {
	if buf.currentX > 0 {
		buf.currentX--
		buf.currentPosition--
	} else if buf.currentLine.prev != nil {
		buf.currentLine = buf.currentLine.prev
		buf.currentY--
		buf.currentX = len(buf.currentLine.bytes) - 1
		buf.currentPosition--
	}
}

func (buf *Buffer) MoveRight() {
	if buf.currentX < len(buf.currentLine.bytes)-1 {
		buf.currentX++
		buf.currentPosition++
	} else if buf.currentLine.next != nil {
		buf.currentLine = buf.currentLine.next
		buf.currentY++
		buf.currentX = 0
		buf.currentPosition++
	}
}

func (buf *Buffer) MoveUp() {
	if buf.currentLine.prev != nil {
		target := sliceLength(buf.currentLine.bytes[:buf.currentX])
		buf.currentPosition -= buf.currentX
		buf.currentLine = buf.currentLine.prev
		buf.currentY--
		buf.currentX = 0
		s := 0
		for _, ch := range buf.currentLine.bytes {
			s += charLength(ch)
			if s > target {
				break
			}
			buf.currentX++
			if s == target {
				break
			}
		}
		buf.currentPosition -= len(buf.currentLine.bytes) - buf.currentX
	}
}

func (buf *Buffer) MoveDown() {
	if buf.currentLine.next != nil {
		target := sliceLength(buf.currentLine.bytes[:buf.currentX])
		buf.currentPosition += len(buf.currentLine.bytes) - buf.currentX
		buf.currentLine = buf.currentLine.next
		buf.currentY++
		buf.currentX = 0
		s := 0
		for _, ch := range buf.currentLine.bytes {
			s += charLength(ch)
			if s > target {
				break
			}
			buf.currentX++
			if s == target {
				break
			}
		}
		buf.currentPosition += buf.currentX
	}
}

func (buf *Buffer) SetPosition(pos int) {
	buf.currentPosition = pos
	buf.currentLine = buf.lines
	buf.currentY = 0
	buf.currentX = 0
	acc := 0
	for {
		if acc+len(buf.currentLine.bytes) > pos {
			buf.currentX = pos - acc
			break
		} else {
			acc += len(buf.currentLine.bytes)
			buf.currentLine = buf.currentLine.next
			buf.currentY++
		}
	}
}

func (buf *Buffer) GetDisplayInformation(screenY, height int) (int, int, int, *Line) {
	if buf.currentY < screenY {
		screenY = buf.currentY
	} else if buf.currentY >= screenY+height {
		screenY = buf.currentY - height + 1
	}
	cursorY := buf.currentY - screenY
	cursorX := sliceLength(buf.currentLine.bytes[:buf.currentX])
	line := buf.currentLine
	for i := 0; i < cursorY; i++ {
		line = line.prev
	}
	return screenY, cursorX, cursorY, line
}

func (buf *Buffer) Lines() *Line {
	return buf.lines
}

func (buf *Buffer) ToString() string {
	builder := bytes.Buffer{}
	lines := buf.lines
	builder.Write(lines.bytes[0:])
	lines = lines.next
	for lines != nil {
		builder.Write(lines.bytes)
		lines = lines.next
	}
	builder.Truncate(builder.Len() - 1)
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
