package gui

import "bytes"

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

func SeqToLines(seq WordSequence, width int) (*Line, int, int) {
	line := &Line{bytes: make([]byte, 0, width+2)}
	startLine := line
	i := 0
	numberOfLines := 1
	numberOfChars := 0
	for {
		word := seq.NextWord()
		if word == nil {
			break
		}
		numberOfChars += len(word)

		if word[0] == '\n' {
			line.bytes = append(line.bytes, '\n')
			lastLine := line
			line = &Line{bytes: make([]byte, 0, width), prev: lastLine}
			lastLine.next = line
			i = 0
			numberOfLines++
			continue
		}

		wordLen := sliceLength(word)
		if i+wordLen <= width {
			line.bytes = append(line.bytes, word...)
			i += wordLen
			continue
		}

		if wordLen <= width {
			lastLine := line
			line = &Line{bytes: make([]byte, 0, width+2), prev: lastLine}
			lastLine.next = line
			line.bytes = append(line.bytes, word...)
			i = wordLen
			numberOfLines++
			continue
		} else {
			line.bytes = append(line.bytes, word[:width-i]...)
			k := width - i
			for k < wordLen {
				lastLine := line
				line = &Line{bytes: make([]byte, 0, width+2), prev: lastLine}
				lastLine.next = line
				i = width
				if wordLen-k < width {
					i = wordLen - k
				}
				line.bytes = append(line.bytes, word[k:k+i]...)
				numberOfLines++
				k += i
			}
			continue
		}
	}
	// a sentinel \n at the end to remove edge cases from various methods
	line.bytes = append(line.bytes, '\n')
	return startLine, numberOfLines, numberOfChars
}

func (buf *Buffer) Resize(width int) {
	buf.lines, buf.numberOfLines, buf.numberOfChars = SeqToLines(NewLineSequence(buf.lines), width)
	buf.SetPosition(buf.currentPosition)
	buf.width = width
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

func (buf *Buffer) findPos(pos int) (int, int, int, *Line) {
	currentPosition := pos
	currentLine := buf.lines
	currentY := 0
	currentX := 0
	acc := 0
	for {
		if acc+len(currentLine.bytes) > pos {
			currentX = pos - acc
			break
		} else {
			acc += len(currentLine.bytes)
			currentLine = currentLine.next
			currentY++
		}
	}
	return currentPosition, currentX, currentY, currentLine
}

func (buf *Buffer) InsertAtCurrent(ch byte) {
	buf.Insert(buf.currentPosition, ch)
}

func (buf *Buffer) Insert(pos int, ch byte) {
	_, currentX, _, currentLine := buf.findPos(pos)
	currentLine.bytes = append(currentLine.bytes, 0)
	copy(currentLine.bytes[currentX+1:], currentLine.bytes[currentX:])
	currentLine.bytes[currentX] = ch
	buf.currentPosition++
	buf.currentX++
}

func (buf *Buffer) DeleteAtCurrent() {
	if buf.numberOfChars > 0 {
		buf.Delete(buf.currentPosition)
	}
}

func (buf *Buffer) BackspaceAtCurrent() {
	if buf.currentPosition > 0 && buf.numberOfChars > 0 {
		buf.Delete(buf.currentPosition - 1)
	}
}

func (buf *Buffer) Delete(pos int) {
	_, currentX, _, currentLine := buf.findPos(pos)
	currentLine.bytes = append(currentLine.bytes[:currentX], currentLine.bytes[currentX+1:]...)
	buf.currentPosition--
	buf.currentX--
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
