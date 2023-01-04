package pass_water_wall

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"math"
	"strings"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/rod_helper"
	"github.com/nfnt/resize"
)

// SimulationTest 模拟滑动过防水墙
func SimulationTest() {
	// 具体的应用见 subhd 的解析器
	// 感谢 https://www.bigs3.com/article/gorod-crack-slider-captcha/
	browser, err := rod_helper.NewBrowserBase(log_helper.GetLogger4Tester(), "", "", false)
	if err != nil {
		println(err.Error())
		return
	}
	defer func() {
		_ = browser.Close()
	}()
	page, err := rod_helper.NewPageNavigate(browser, "https://007.qq.com/online.html", 10*time.Second)
	if err != nil {
		println(err.Error())
		return
	}
	defer func() {
		_ = page.Close()
	}()
	// 切换到可疑用户
	page.MustElement("#app > section.wp-on-online > div > div > div > div.wp-on-box.col-md-5.col-md-offset-1 > div.wp-onb-tit > a:nth-child(2)").MustClick()
	//模擬Click點擊 "體驗驗證碼" 按鈕
	page.MustElement("#code").MustClick()
	//等待驗證碼窗體載入
	page.MustElement("#tcaptcha_iframe").MustWaitLoad()
	//進入到iframe
	iframe := page.MustElement("#tcaptcha_iframe").MustFrame()
	//等待拖動條加載, 延遲500秒檢測變化, 以確認加載完畢
	iframe.MustElement("#tcaptcha_drag_button").WaitStable(500 * time.Millisecond)
	//等待缺口圖像載入
	iframe.MustElement("#slideBg").MustWaitLoad()

	//取得帶缺口圖像
	shadowbg := iframe.MustElement("#slideBg").MustResource()
	//取得原始圖像
	src := iframe.MustElement("#slideBg").MustProperty("src")
	fullbg, fileName, err := pkg.DownFile(log_helper.GetLogger4Tester(), strings.Replace(src.String(), "img_index=1", "img_index=0", 1))
	if err != nil {
		return
	}
	println(fileName)
	//取得img展示的真實尺寸
	bgbox := iframe.MustElement("#slideBg").MustShape().Box()
	height, width := uint(math.Round(bgbox.Height)), uint(math.Round(bgbox.Width))
	//裁剪圖像
	shadowbg_img, _ := jpeg.Decode(bytes.NewReader(shadowbg))
	shadowbg_img = resize.Resize(width, height, shadowbg_img, resize.Lanczos3)
	fullbg_img, _ := jpeg.Decode(bytes.NewReader(fullbg))
	fullbg_img = resize.Resize(width, height, fullbg_img, resize.Lanczos3)

	//啓始left，排除干擾部份，所以右移10個像素
	left := fullbg_img.Bounds().Min.X + 10

	//啓始top, 排除干擾部份, 所以下移10個像素
	top := fullbg_img.Bounds().Min.Y + 10

	//最大left, 排除干擾部份, 所以左移10個像素
	maxleft := fullbg_img.Bounds().Max.X - 10

	//最大top, 排除干擾部份, 所以上移10個像素
	maxtop := fullbg_img.Bounds().Max.Y - 10

	//rgb比较阈值, 超出此阈值及代表找到缺口位置
	threshold := 20

	//缺口偏移, 拖動按鈕初始會偏移27.5
	distance := -27.5

	//取絕對值方法
	abs := func(n int) int {
		if n < 0 {
			return -n
		}
		return n
	}
search:
	for i := left; i <= maxleft; i++ {
		for j := top; j <= maxtop; j++ {
			color_a_R, color_a_G, color_a_B, _ := fullbg_img.At(i, j).RGBA()
			color_b_R, color_b_G, color_b_B, _ := shadowbg_img.At(i, j).RGBA()
			color_a_R, color_a_G, color_a_B = color_a_R>>8, color_a_G>>8, color_a_B>>8
			color_b_R, color_b_G, color_b_B = color_b_R>>8, color_b_G>>8, color_b_B>>8
			if abs(int(color_a_R)-int(color_b_R)) > threshold ||
				abs(int(color_a_G)-int(color_b_G)) > threshold ||
				abs(int(color_a_B)-int(color_b_B)) > threshold {
				distance += float64(i)
				fmt.Printf("info: 對比完畢, 偏移量: %v\n", distance)
				break search
			}
		}
	}

	//獲取拖動按鈕形狀
	dragbtnbox := iframe.MustElement("#tcaptcha_drag_thumb").MustShape().Box()
	//启用滑鼠功能
	mouse := page.Mouse
	//模擬滑鼠移動至拖動按鈕處, 右移3的原因: 拖動按鈕比滑塊圖大3個像素
	mouse.MustMove(dragbtnbox.X+3, dragbtnbox.Y+(dragbtnbox.Height/2))
	//按下滑鼠左鍵
	mouse.MustDown("left")
	//開始拖動
	mouse.Move(dragbtnbox.X+distance, dragbtnbox.Y+(dragbtnbox.Height/2), 20)
	//鬆開滑鼠左鍵, 拖动完毕
	mouse.MustUp("left")
	//截圖保存
	page.MustScreenshot("result.png")
}
