package sub_parser_hub

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
)

type SubParserHub struct {
	*sub_parser_hub.SubParserHub
}

func NewSubParserHub() *SubParserHub {

	return &SubParserHub{sub_parser_hub.NewSubParserHub(nil, ass.NewParser(nil), srt.NewParser(nil))}
}
