package random_useragent

import (
	browser "github.com/allanpk716/fake-useragent"
	"math/rand"
	"time"
)

func RandomUserAgent(UserOrSearchEngine bool) string {
	if UserOrSearchEngine == true {
		return browser.Random()
	} else {
		// From https://www.cnblogs.com/gengyufei/p/12641200.html
		return engineUAList[random.Intn(len(engineUAList))]
	}
}

var (
	random       = rand.New(rand.NewSource(time.Now().UnixNano()))
	engineUAList = []string{
		// 百度搜索User-Agent：
		// 百度 PC UA
		"Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)",
		"Mozilla/5.0 (compatible; Baiduspider-render/2.0; +http://www.baidu.com/search/spider.html)",
		// 百度移动 UA
		"Mozilla/5.0 (Linux;u;Android 4.2.2;zh-cn;) AppleWebKit/534.46 (KHTML,like Gecko) Version/5.1",
		"Mobile Safari/10600.6.3 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1 (compatible; Baiduspider-render/2.0; +http://www.baidu.com/search/spider.html)",
		// 百度图片UA
		//"Baiduspider-image+(+http://www.baidu.com/search/spider.htm)",
		// 神马搜索User-Agent：
		// B神马搜索 PC UA
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.81 YisouSpider/5.0 Safari/537.36",
		// 神马搜索移动 UA
		"Mozilla/5.0 (iPhone; CPU iPhone OS 10_3 like Mac OS X) AppleWebKit/602.1.50 (KHTML, like Gecko) CriOS/56.0.2924.75 Mobile/14E5239e YisouSpider/5.0 Safari/602.1",
		// 谷歌User-Agent：
		// 谷歌 PC UA
		"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		// 谷歌移动UA
		"AdsBot-Google-Mobile (+http://www.google.com/mobile/adsbot.html) Mozilla (iPhone; U; CPU iPhone OS 3 0 like Mac OS X) AppleWebKit (KHTML, like Gecko) Mobile Safari",
		// 谷歌图片UA
		"Mozilla/5.0 (compatible; Googlebot-Image/1.0; +http://www.google.com/bot.html)",
		// 搜狗User-Agent：
		// 搜索 PC UA
		"Sogou web spider/4.0(+http://www.sogou.com/docs/help/webmasters.htm#07)",
		// 搜狗图片 UA
		"Sogou Pic Spider/3.0(+http://www.sogou.com/docs/help/webmasters.htm#07)",
		// 搜狗新闻UA
		"Sogou News Spider/4.0(+http://www.sogou.com/docs/help/webmasters.htm#07)",
		// 搜狗视频UA
		"Sogou Video Spider/3.0(+http://www.sogou.com/docs/help/webmasters.htm#07)",
		// 360搜索User-Agent：
		// 360搜索UA
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0);",
		// 360移动UA
		"Mozilla/5.0 (Linux; U; Android 4.0.2; en-us; Galaxy Nexus Build/ICL53F) AppleWebKit/534.30 (KHTML, like Gecko)Version/4.0 Mobile Safari/534.30; 360Spider",
		"Mozilla/5.0 (Linux; U; Android 4.0.2; en-us; Galaxy Nexus Build/ICL53F) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/534.30; HaosouSpider",
		// 360安全UA
		"360spider (http://webscan.360.cn)",
		// 必应User-Agent：
		"Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
		// 搜搜User-Agent：
		// 搜搜UA：
		"Sosospider+(+http://help.soso.com/webspider.htm)",
		// 搜搜图片UA：
		"Sosoimagespider+(+http://help.soso.com/soso-image-spider.htm)",
		// 雅虎User-Agent：
		// 雅虎中文UA：
		"Mozilla/5.0 (compatible; Yahoo! Slurp China; http://misc.yahoo.com.cn/help.html)",
		// 雅虎英文UA：
		"Mozilla/5.0 (compatible; Yahoo! Slurp; http://help.yahoo.com/help/us/ysearch/slurp)",
	}
)
