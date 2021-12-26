package sub_timeline_fixer

import (
	"fmt"
	"testing"
)

func TestPipeline_getFramerateRatios2Try(t *testing.T) {

	outList := NewPipeline().getFramerateRatios2Try()
	for i, value := range outList {
		println(i, fmt.Sprintf("%v", value))
	}
}
