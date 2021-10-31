package language

import (
	"github.com/go-creed/sat"
	"github.com/saintfish/chardet"
)

var (
	chDict   = sat.DefaultDict()
	detector = chardet.NewTextDetector()
)
