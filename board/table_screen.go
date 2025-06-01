package board

import "github.com/ascii-arcade/farkle/screen"

type tableScreen struct {
	model *Model

	rollTickCount int
	rolling       bool
}

func (s *tableScreen) WithModel(model any) screen.Screen {
	s.model = model.(*Model)
	return s
}
