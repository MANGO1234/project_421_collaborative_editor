package gui

type WordSequence interface {
	NextWord() []byte
}

type LineSequence struct {
	lines    *Line
	currentX int
}

func NewLineSequence(lines *Line) *LineSequence {
	return &LineSequence{lines: lines, currentX: 0}
}

func (seq *LineSequence) NextWord() []byte {
	for seq.lines != nil && seq.currentX == len(seq.lines.bytes) {
		seq.lines = seq.lines.next
		seq.currentX = 0
	}

	if seq.lines.next == nil {
		return nil
	}

	ch := seq.lines.bytes[seq.currentX]
	if ch == ' ' {
		seq.currentX++
		return []byte(" ")
	}
	if ch == '\t' {
		seq.currentX++
		return []byte("\t")
	}
	if ch == '\n' {
		seq.currentX++
		return []byte("\n")
	}

	start := seq.currentX
	end := seq.currentX + 1
	multiline := true
	for end < len(seq.lines.bytes) {
		ch := seq.lines.bytes[end]
		if ch == ' ' || ch == '\t' || ch == '\n' {
			multiline = false
			break
		}
		end++
	}

	if !multiline {
		seq.currentX = end
		return seq.lines.bytes[start:end]
	} else {
		chs := make([]byte, 0)
		chs = append(chs, seq.lines.bytes[start:end]...)
		for {
			if seq.lines.next == nil {
				break
			}
			ch := seq.lines.next.bytes[0]
			if ch == ' ' || ch == '\t' || ch == '\n' {
				break
			}
			seq.lines = seq.lines.next
			end = 1
			for end < len(seq.lines.bytes) {
				ch := seq.lines.bytes[end]
				if ch == ' ' || ch == '\t' || ch == '\n' {
					break
				}
				end++
			}
			chs = append(chs, seq.lines.bytes[:end]...)
			if end != len(seq.lines.bytes) {
				break
			}
		}
		seq.currentX = end
		return chs
	}
}

type StringSequence struct {
	str string
	pos int
}

func NewStringSequence(str string) *StringSequence {
	return &StringSequence{str, 0}
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

// **********************************************
// ******************** Util ********************
// **********************************************
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
