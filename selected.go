package textcell

type selectedText struct {
	lines          []*selectedLnArea
	pivotX, pivotY int
	dir            byte // 'R' or 'L'
}

type selectedLnArea struct {
	start, end int
	y          int
}

func newSelectedText(pivotX, pivotY int) *selectedText {
	return &selectedText{
		lines:  make([]*selectedLnArea, 0, 8),
		dir:    0,
		pivotX: pivotX,
		pivotY: pivotY,
	}
}

// calcDirLnCount returns direction of selection and
// the number of lines with selected text on it.
func (s *selectedText) calcDirLnCount(cx, cy int) (byte, int) {
	if s.pivotY == cy {
		if s.pivotX == cx {
			return 0, 0
		}
		if s.pivotX < cx {
			return 'R', 1
		} else {
			return 'L', 1
		}
	}
	if v := s.pivotY - cy; v < 0 {
		return 'R', (v * -1) + 1
	} else {
		return 'L', v + 1
	}
}
