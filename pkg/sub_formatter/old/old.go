package old

import (
	"path/filepath"
	"strings"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	language2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/language"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/subparser"
)

/*
	整个是早期版本使用的字幕命名格式，现在已经弃用，通过 HotFix01 进行调整过。
	后续是无需关心的
*/

// IsOldVersionSubPrefixName 是否是老版本的字幕命名 .chs_en[shooter] ，符合也返回这个部分＋字幕格式后缀名 .chs_en[shooter].ass, 修改后的名称
func IsOldVersionSubPrefixName(subFileName string) (bool, string, string) {
	/*
		{
			name: "chs_en[shooter]", args: args{
			subFileName: "Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chs_en[shooter].ass"},
			want: true,
			want1: ".chs_en[shooter].ass",
			want2: "Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chinese(简英,shooter).ass"
		},
			传入的必须是字幕格式的文件，这个就再之前判断，不要在这里再判断
			传入的文件名可能有一下几种情况
			无罪之最 - S01E01 - 重建生活.chs[shooter].ass
			无罪之最 - S01E03 - 初见端倪.zh.srt
			Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chs_en.ass
			那么就需要先剔除，字幕的格式后缀名，然后再向后取后缀名就是 .chs[shooter] or .zh
			再判断即可
	*/
	// 无罪之最 - S01E01 - 重建生活.chs[shooter].ass -> 无罪之最 - S01E01 - 重建生活.chs[shooter]
	subTypeExt := filepath.Ext(subFileName)
	subFileNameWithOutExt := strings.ReplaceAll(subFileName, subTypeExt, "")
	// .chs[shooter]
	nowExt := filepath.Ext(subFileNameWithOutExt)
	// .chs_en[shooter].ass
	orgMixExt := nowExt + subTypeExt
	orgFileNameWithOutOrgMixExt := strings.ReplaceAll(subFileName, orgMixExt, "")
	// 这里也有两种情况，一种是单字幕 SaveMultiSub: false
	// 一种的保存了多字幕 SaveMultiSub: true
	// 先判断 单字幕
	switch nowExt {
	case language2.Emby_chs:
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, language2.MatchLangChs, subTypeExt, "", true)
	case language2.Emby_cht:
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, language2.MatchLangCht, subTypeExt, "", false)
	case language2.Emby_chs_en:
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, language2.MatchLangChsEn, subTypeExt, "", true)
	case language2.Emby_cht_en:
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, language2.MatchLangChtEn, subTypeExt, "", false)
	case language2.Emby_chs_jp:
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, language2.MatchLangChsJp, subTypeExt, "", true)
	case language2.Emby_cht_jp:
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, language2.MatchLangChtJp, subTypeExt, "", false)
	case language2.Emby_chs_kr:
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, language2.MatchLangChsKr, subTypeExt, "", true)
	case language2.Emby_cht_kr:
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, language2.MatchLangChtKr, subTypeExt, "", false)
	}
	// 再判断 多字幕情况
	spStrings := strings.Split(nowExt, "[")
	if len(spStrings) != 2 {
		return false, "", ""
	}
	// 分两段来判断是否符合标准
	// 第一段
	firstOk := true
	lang := language2.MatchLangChs
	site := ""
	switch spStrings[0] {
	case language2.Emby_chs:
		lang = language2.MatchLangChs
	case language2.Emby_cht:
		lang = language2.MatchLangCht
	case language2.Emby_chs_en:
		lang = language2.MatchLangChsEn
	case language2.Emby_cht_en:
		lang = language2.MatchLangChtEn
	case language2.Emby_chs_jp:
		lang = language2.MatchLangChsJp
	case language2.Emby_cht_jp:
		lang = language2.MatchLangChtJp
	case language2.Emby_chs_kr:
		lang = language2.MatchLangChsKr
	case language2.Emby_cht_kr:
		lang = language2.MatchLangChtKr
	default:
		firstOk = false
	}
	// 第二段
	secondOk := true
	tmpSecond := strings.ReplaceAll(spStrings[1], "]", "")
	switch tmpSecond {
	case common.SubSiteZiMuKu:
		site = common.SubSiteZiMuKu
	case common.SubSiteSubHd:
		site = common.SubSiteSubHd
	case common.SubSiteShooter:
		site = common.SubSiteShooter
	case common.SubSiteXunLei:
		site = common.SubSiteXunLei
	default:
		secondOk = false
	}
	// 都要符合条件
	if firstOk == true && secondOk == true {
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, lang, subTypeExt, site, false)
	}
	return false, "", ""
}

func makeMixSubExtString(orgFileNameWithOutExt, lang string, ext, site string, beDefault bool) string {

	tmpDefault := ""
	if beDefault == true {
		tmpDefault = subparser.Sub_Ext_Mark_Default
	}

	if site == "" {
		return orgFileNameWithOutExt + language2.Emby_chinese + "(" + lang + ")" + tmpDefault + ext
	}
	return orgFileNameWithOutExt + language2.Emby_chinese + "(" + lang + "," + site + ")" + tmpDefault + ext
}
