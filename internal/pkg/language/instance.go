package language

import (
	"github.com/go-creed/sat"
	"github.com/saintfish/chardet"
)

var (
	ChDict   = sat.DefaultDict()
	detector = chardet.NewTextDetector()
)
