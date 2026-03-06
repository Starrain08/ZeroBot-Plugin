package gif

import (
	"errors"
	"image/color"
	"math/rand"
	"os"
	"strconv"
	"sync"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
)

// pa 爬
func pa(cc *context, args ...string) (string, error) {
	_ = args
	tou, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	// 随机爬图序号
	rand := rand.Intn(92) + 1
	if file.IsNotExist(datapath + "materials/pa") {
		err = os.MkdirAll(datapath+"materials/pa", 0755)
		if err != nil {
			return "", err
		}
	}
	f, err := dlblock("pa/" + strconv.Itoa(rand) + ".png")
	if err != nil {
		return "", err
	}
	imgf, err := imgfactory.LoadFirstFrame(f, 0, 0)
	if err != nil {
		return "", err
	}
	base64Bytes, err := imgfactory.ToBase64(imgf.InsertUp(tou, 100, 100, 0, 400).Image())
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// si 撕
func si(cc *context, args ...string) (string, error) {
	_ = args
	tou, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	im1 := imgfactory.Rotate(tou, 20, 380, 380)
	im2 := imgfactory.Rotate(tou, -12, 380, 380)
	if file.IsNotExist(datapath + "materials/si") {
		err = os.MkdirAll(datapath+"materials/si", 0755)
		if err != nil {
			return "", err
		}
	}
	f, err := dlblock("si/0.png")
	if err != nil {
		return "", err
	}
	imgf, err := imgfactory.LoadFirstFrame(f, 0, 0)
	if err != nil {
		return "", err
	}
	base64Bytes, err := imgfactory.ToBase64(imgf.InsertBottom(im1.Image(), im1.W(), im1.H(), -3, 370).InsertBottom(im2.Image(), im2.W(), im2.H(), 653, 310).Image())
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// flipV 上翻,下翻
func flipV(cc *context, args ...string) (string, error) {
	_ = args
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.FlipV().Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// flipH 左翻,右翻
func flipH(cc *context, args ...string) (string, error) {
	_ = args
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.FlipH().Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// invert 反色
func invert(cc *context, args ...string) (string, error) {
	_ = args
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.Invert().Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// blur 反色
func blur(cc *context, args ...string) (string, error) {
	_ = args
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.Blur(10).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// grayscale 灰度
func grayscale(cc *context, args ...string) (string, error) {
	_ = args
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.Grayscale().Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// invertAndGrayscale 负片
func invertAndGrayscale(cc *context, args ...string) (string, error) {
	_ = args
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.Invert().Grayscale().Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// convolve3x3 浮雕
func convolve3x3(cc *context, args ...string) (string, error) {
	_ = args
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.Relief().Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// rotate 旋转
func rotate(cc *context, args ...string) (string, error) {
	_ = args
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	r, _ := strconv.ParseFloat(args[0], 64)
	imgnrgba := imgfactory.Rotate(im.Image(), r, 0, 0).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// deformation 变形
func deformation(cc *context, args ...string) (string, error) {
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	w, err := strconv.Atoi(args[0])
	if err != nil {
		return "", err
	}
	h, err := strconv.Atoi(args[1])
	if err != nil {
		return "", err
	}
	imgnrgba := imgfactory.Size(im.Image(), w, h).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// alike 你像个xxx一样
func alike(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("alike", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 82, 69)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertUp(im.Image(), 0, 0, 136, 21).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// marriage
func marriage(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("marriage", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 1080, 1080)
	if err != nil {
		return "", err
	}
	imgnrgba := im.InsertUp(imgs[0].Image(), 0, 0, 0, 0).InsertUp(imgs[1].Image(), 0, 0, 800, 0).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// anyasuki 阿尼亚喜欢
func anyasuki(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("anyasuki", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(475, 540)
	canvas.DrawImage(imgfactory.Size(face, 347, 267).Image(), 82, 53)
	canvas.DrawImage(back, 0, 0)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 30); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "阿尼亚喜欢这个"
	}
	l, _ := canvas.MeasureString(args[0])
	if l > 500 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (500-l)/2.0, 535)
	base64Bytes, err := imgfactory.ToBase64(canvas.Image())
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// alwaysLike 我永远喜欢
func alwaysLike(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("always_like", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(830, 599)
	canvas.DrawImage(back, 0, 0)
	canvas.DrawImage(imgfactory.Size(face, 380, 380).Image(), 44, 74)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 56); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "你们"
	}
	args[0] = "我永远喜欢" + args[0]
	l, _ := canvas.MeasureString(args[0])
	if l > 830 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (830-l)/2.0, 559)
	base64Bytes, err := imgfactory.ToBase64(canvas.Image())
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// decentKiss 像样的亲亲
func decentKiss(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("decent_kiss", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 589, 577)
	if err != nil {
		return "", err
	}
	imgnrgba := im.InsertUp(imgs[0].Image(), 0, 0, 0, 0).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// chinaFlag 国旗
func chinaFlag(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("china_flag", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 410, 410)
	if err != nil {
		return "", err
	}
	imgnrgba := im.InsertUp(imgs[0].Image(), 0, 0, 0, 0).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// dontTouch 不要靠近
func dontTouch(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("dont_touch", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 410, 410)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertUp(im.Image(), 148, 148, 46, 238).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// universal 万能表情 空白表情
func universal(cc *context, args ...string) (string, error) {
	_ = args
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(500, 550)
	canvas.DrawImage(imgfactory.Size(face, 500, 500).Image(), 0, 0)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 40); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "在此处添加文字"
	}
	l, _ := canvas.MeasureString(args[0])
	if l > 500 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (500-l)/2.0, 545)
	base64Bytes, err := imgfactory.ToBase64(canvas.Image())
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// interview 采访
func interview(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("interview", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	huaji, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	microphone, err := gg.LoadImage(c[1])
	if err != nil {
		return "", err
	}
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(600, 300)
	canvas.DrawImage(imgfactory.Size(face, 124, 124).Image(), 100, 50)
	canvas.DrawImage(huaji, 376, 50)
	canvas.DrawImage(microphone, 300, 50)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 40); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "采访大佬经验"
	}
	l, _ := canvas.MeasureString(args[0])
	if l > 600 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (600-l)/2.0, 270)
	base64Bytes, err := imgfactory.ToBase64(canvas.Image())
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// need 需要 你可能需要
func need(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("need", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 114, 114)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 0, 0, 327, 232).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// paint 这像画吗
func paint(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("paint", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 117, 135)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(imgfactory.Rotate(im.Image(), 4, 0, 0).Image(), 0, 0, 95, 107).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// painter 小画家
func painter(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("painter", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 240, 345)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 0, 0, 125, 91).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// perfect 完美
func perfect(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("perfect", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 310, 460)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertUp(im.Image(), 0, 0, 313, 64).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// playGame 玩游戏
func playGame(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("play_game", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(526, 503)
	canvas.DrawImage(imgfactory.Rotate(face, 10, 225, 160).Image(), 161, 117)
	canvas.DrawImage(back, 0, 0)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 40); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "来玩休闲游戏啊"
	}
	l, _ := canvas.MeasureString(args[0])
	if l > 526 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (526-l)/2.0, 483)
	base64Bytes, err := imgfactory.ToBase64(canvas.Image())
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// police 出警
func police(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("police", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 245, 245)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 0, 0, 224, 46).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// police1 警察
func police1(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("police", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 60, 75)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[1].InsertBottom(imgfactory.Rotate(im.Image(), 16, 0, 0).Image(), 0, 0, 37, 291).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// prpr 舔 舔屏 prpr
func prpr(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("prpr", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 330, 330)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(imgfactory.Rotate(im.Image(), 8, 0, 0).Image(), 0, 0, 46, 264).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// safeSense 安全感
func safeSense(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("safe_sense", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(430, 478)
	canvas.DrawImage(back, 0, 0)
	canvas.DrawImage(imgfactory.Size(face, 215, 343).Image(), 215, 135)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 30); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "你给我的安全感远不如他的万分之一"
	}

	l, _ := canvas.MeasureString(args[0])
	if l > 860 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0][:len(args[0])/2], (430-l/2)/2.0, 40)
	canvas.DrawString(args[0][len(args[0])/2:], (430-l/2)/2.0, 80)
	base64Bytes, err := imgfactory.ToBase64(canvas.Image())
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// support 精神支柱
func support(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("support", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 815, 815)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(imgfactory.Rotate(im.Image(), 23, 0, 0).Image(), 0, 0, -172, -17).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// thinkwhat 想什么
func thinkwhat(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("thinkwhat", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 534, 493)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 0, 0, 530, 0).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// wallpaper 墙纸
func wallpaper(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("wallpaper", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 775, 496)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 0, 0, 260, 580).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// whyatme 为什么at我
func whyatme(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("whyatme", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 265, 265)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 0, 0, 42, 13).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// makeFriend 交个朋友
func makeFriend(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("make_friend", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(1000, 1000)
	canvas.DrawImage(imgfactory.Size(face, 1000, 1000).Image(), 0, 0)
	canvas.DrawImage(imgfactory.Rotate(face, 9, 250, 250).Image(), 743, 845)
	canvas.DrawImage(imgfactory.Rotate(face, 9, 55, 55).Image(), 836, 722)
	canvas.DrawImage(back, 0, 0)
	canvas.SetColor(color.White)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 20); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "我"
	}
	l, _ := canvas.MeasureString(args[0])
	if l > 230 {
		return "", errors.New("文字消息太长了")
	}
	canvas.Rotate(gg.Radians(-9))
	canvas.DrawString(args[0], 595, 819)
	base64Bytes, err := imgfactory.ToBase64(canvas.Image())
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// backToWork 打工人, 继续干活
func backToWork(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("back_to_work", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 220, 310)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(imgfactory.Rotate(im.Image(), 25, 0, 0).Image(), 0, 0, 56, 32).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// coupon 兑换券
func coupon(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("coupon", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "群主陪睡券"
	}
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(500, 355)
	canvas.DrawImage(back, 0, 0)
	canvas.Rotate(gg.Radians(-22))
	canvas.DrawImage(imgfactory.Size(face, 60, 60).Image(), 100, 163)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 30); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "陪睡券"
	}
	l, _ := canvas.MeasureString(args[0])
	if l > 270 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawStringAnchored(args[0], 135, 255, 0.5, 0.5)
	canvas.DrawStringAnchored("（永久有效）", 135, 295, 0.5, 0.5)
	base64Bytes, err := imgfactory.ToBase64(canvas.Image())
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// distracted 注意力涣散
func distracted(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("distracted", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 500, 500)
	if err != nil {
		return "", err
	}
	imgnrgba := im.InsertUp(imgs[0].Image(), 0, 0, 140, 320).InsertUp(imgs[1].Image(), 0, 0, 0, 0).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// throw 扔
func throw(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("throw", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	face, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertUpC(imgfactory.Rotate(face, float64(rand.Intn(360)), 143, 143).Image(), 0, 0, 86, 249).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// 远离
func yuanli(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("yuanli", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 534, 493)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 420, 420, 45, 90).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// 不是你老婆
func nowife(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("nowife", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 534, 493)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 400, 400, 112, 81).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// youer 你老婆
func youer(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("youer", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	tou, err := cc.getLogo(120, 120)
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(690, 690)
	canvas.DrawImage(back, 0, 0)
	canvas.DrawImage(imgfactory.Size(tou, 350, 350).Image(), 55, 165)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 56); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "老婆真棒"
	}
	args[0] = "你的" + args[0]
	l, _ := canvas.MeasureString(args[0])
	if l > 830 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (830-l)/3.0, 630)
	base64Bytes, err := imgfactory.ToBase64(canvas.Image())
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// xiaotiamshi 小天使
func xiaotianshi(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("xiaotianshi", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(522, 665)
	canvas.DrawImage(back, 0, 0)
	canvas.DrawImage(imgfactory.Size(face, 480, 480).Image(), 20, 80)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 35); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "我老婆"
	}
	args[0] = "请问你们看到" + args[0] + "了吗？"
	l, _ := canvas.MeasureString(args[0])
	if l > 830 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (830-l)/10, 50)
	base64Bytes, err := imgfactory.ToBase64(canvas.Image())
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// 不要再看这些了
func neko(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("neko", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 712, 949)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(imgfactory.Rotate(im.Image(), 0, 0, 0).Image(), 450, 450, 0, 170).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// 给我变
func bian(cc *context, args ...string) (string, error) {
	_ = args
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("bian", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 640, 550)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(imgfactory.Rotate(im.Image(), 0, 0, 0).Image(), 380, 380, 225, -20).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// van 玩一下
func van(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("van", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(522, 665)
	canvas.DrawImage(back, 0, 0)
	canvas.DrawImage(imgfactory.Size(face, 480, 480).Image(), 20, 80)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 35); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "RBQ"
	}
	args[0] = "请问你们看到" + args[0] + "了吗？"
	l, _ := canvas.MeasureString(args[0])
	if l > 830 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (830-l)/10, 50)
	base64Bytes, err := imgfactory.ToBase64(canvas.Image())
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// eihei 诶嘿
func eihei(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("eihei", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 690, 690)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 450, 450, 121, 162).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// fanfa 犯法
func fanfa(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("fanfa", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	face, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	m1 := imgfactory.Rotate(face, 45, 110, 110)
	imgnrgba := imgs[0].InsertUp(m1.Image(), 0, 0, 125, 360).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// huai 怀
func huai(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("huai", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 640, 640)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 640, 640, 0, 0).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// haowan 好玩
func haowan(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("haowan", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	face, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(face, 90, 90, 321, 172).Image()
	base64Bytes, err := imgfactory.ToBase64(imgnrgba)
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}

// mengbi 蒙蔽
func mengbi(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("mengbi", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(1080, 1080)
	canvas.DrawImage(back, 0, 0)
	canvas.DrawImage(imgfactory.Size(face, 100, 100).Image(), 392, 460)
	canvas.DrawImage(imgfactory.Size(face, 100, 100).Image(), 606, 443)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 80); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = ""
	}
	l, _ := canvas.MeasureString(args[0])
	if l > 1080 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (1080-l)/2, 1000)
	base64Bytes, err := imgfactory.ToBase64(canvas.Image())
	if err != nil {
		return "", err
	}
	return "base64://" + binary.BytesToString(base64Bytes), nil
}
