package charset

import (
	"fmt"
	"golang.org/x/net/html/charset"
	"testing"
)

func TestConvert(t *testing.T) {

	gbk := []byte{206, 210, 202, 199, 71, 66, 75}
	encoding, name, certain := charset.DetermineEncoding(gbk, "text/html")
	fmt.Printf("编码：%v\n名称：%s\n确定：%t\n", encoding, name, certain)
}
