// Package tupian 图片获取集合
package tupian

import (
	"github.com/FloatTech/AnimeAPI/bilibili"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	defaultReferer = "https://weibo.com/"
)

var (
	imageTypeURLs = map[string]string{
		"兽耳":    "https://iw233.cn/api.php?sort=cat&referer",
		"白丝":    "http://aikohfiosehgairl.fgimax2.fgnwctvip.com/uyfvnuvhgbuiesbrghiuudvbfkllsgdhngvbhsdfklbghdfsjksdhnvfgkhdfkslgvhhrjkdshgnverhbgkrthbklg.php?sort=ergbskjhebrgkjlhkerjsbkbregsbg",
		"黑丝":    "http://aikohfiosehgairl.fgimax2.fgnwctvip.com/uyfvnuvhgbuiesbrghiuudvbfkllsgdhngvbhsdfklbghdfsjksdhnvfgkhdfkslgvhhrjkdshgnverhbgkrthbklg.php?sort=rsetbgsekbjlghelkrabvfgheiv",
		"丝袜":    "http://aikohfiosehgairl.fgimax2.fgnwctvip.com/uyfvnuvhgbuiesbrghiuudvbfkllsgdhngvbhsdfklbghdfsjksdhnvfgkhdfkslgvhhrjkdshgnverhbgkrthbklg.php?sort=dsrgvkbaergfvyagvbkjavfwe",
		"随机壁纸":  "https://iw233.cn/api.php?sort=iw233",
		"白毛":    "https://iw233.cn/api.php?sort=yin",
		"星空":    "https://iw233.cn/api.php?sort=xing",
		"涩涩达咩":  "https://sex.nyan.xyz/api/v2/img?r18",
		"我要涩涩":  "https://sex.nyan.xyz/api/v2/img?r18",
		"随机表情包": "https://iw233.cn/api.php?sort=img",
		"cos":   "http://aikohfiosehgairl.fgimax2.fgnwctvip.com/uyfvnuvhgbuiesbrghiuudvbfkllsgdhngvbhsdfklbghdfsjksdhnvfgkhdfkslgvhhrjkdshgnverhbgkrthbklg.php/?sort=cos",
		"盲盒":    "https://iw233.cn/api.php?sort=random",
		"开盲盒":   "https://iw233.cn/api.php?sort=random",
	}
)

func init() {
	engine := control.Register("tupian", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "全部图片指令\n" +
			"- cos\n" +
			"- 兽耳\n" +
			"- 白毛\n" +
			"- 黑丝\n" +
			"- 白丝\n" +
			"- 丝袜\n" +
			"- 星空\n" +
			"- 开盲盒\n" +
			"- 随机壁纸\n" +
			"- 随机表情包\n" +
			"- 涩涩达咩/我要涩涩\n",
	})
	engine.OnFullMatchGroup([]string{"随机壁纸", "兽耳", "星空", "白毛", "我要涩涩", "涩涩达咩", "白丝", "黑丝", "丝袜", "随机表情包", "cos", "盲盒", "开盲盒"}).SetBlock(true).
		Handle(handleImageRequest)
}

func handleImageRequest(ctx *zero.Ctx) {
	imageType := ctx.State["matched"].(string)
	url, ok := imageTypeURLs[imageType]
	if !ok {
		ctx.SendChain(message.Text("未找到该类型的图片"))
		return
	}

	data, err := fetchImageData(url)
	if err != nil {
		ctx.SendChain(message.Text(err.Error()))
		return
	}

	ctx.SendChain(message.ImageBytes(data))
}

func fetchImageData(url string) ([]byte, error) {
	realURL, err := bilibili.GetRealURL(url)
	if err != nil {
		return nil, err
	}

	data, err := web.RequestDataWith(web.NewDefaultClient(), realURL, "", defaultReferer, "", nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}
