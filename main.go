// check coin price and then notify by send sms
package main

import (
	"os"
	"fmt"
	"time"
	"math"
	"strings"
	"strconv"
	"syscall"
	"net/http"
	"os/signal"

	"github.com/urfave/cli"
	"github.com/yudai/gotty/pkg/homedir"
	"github.com/smileboywtu/CoinNotify/aliyun"
	"github.com/smileboywtu/CoinNotify/common"
	"github.com/smileboywtu/CoinNotify/feixiaohao"
)

type TaskContext struct {
	LastNotifyTime map[string]int64
	LastRecord     map[string]float32
	Cookies        []*http.Cookie
	Filter         feixiaohao.CoinFilter
	AliyunCtx      aliyun.AliyunSMSOpt
}

// RenewCookies renew feixiaohao cookies
func RenewCookies(cookies []*http.Cookie, meta feixiaohao.UserLoginMeta) chan struct{} {

	// one day renew
	ticker := time.NewTicker(3600 * 12 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				cookie, _ := feixiaohao.Login(meta)
				copy(cookies, cookie)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	return quit
}

func Task(ctx *TaskContext, errc chan error) {
	pricemeta, err := feixiaohao.GetUserTicket(ctx.Cookies, ctx.Filter)
	if err != nil {
		go func() {
			errc <- err
		}()
	}

	for _, meta := range pricemeta {

		notify, percentf := NeedNotify(meta, *ctx)
		if notify {
			errs := aliyun.SendSMS(ctx.AliyunCtx, aliyun.SMSContentCtx{
				meta.Platform,
				meta.CoinType,
				meta.Price,
				meta.Percent,
			})
			if errs != nil {
				go func() {
					errc <- err
				}()
			}

			ctx.LastNotifyTime[meta.CoinType] = time.Now().Unix()
		}

		ctx.LastRecord[meta.CoinType] = percentf
	}
}

func NeedNotify(meta feixiaohao.CoinPriceMeta, ctx TaskContext) (bool, float32) {

	percentf, errs := ConvertPercent2Float(meta.Percent)
	if errs != nil {
		return false, 0.0
	}

	// time limit
	if ctx.LastNotifyTime[meta.CoinType] == 0 || (ctx.LastNotifyTime[meta.CoinType] > 0 && time.Now().Unix()-ctx.LastNotifyTime[meta.CoinType] >= ctx.Filter.TimePeriod) {
		return true, percentf
	}

	//  check if price reach threshold
	if float32(percentf) >= ctx.Filter.High || float32(percentf) <= ctx.Filter.Low {
		return true, percentf
	}

	// amplitude
	if math.Abs(float64(ctx.LastRecord[meta.CoinType]-percentf)) >= float64(ctx.Filter.Amplitude) {
		return true, percentf
	}

	return false, percentf
}

func ConvertPercent2Float(percent string) (float32, error) {
	// strip space
	percent = strings.TrimSpace(percent)

	// trim %
	percent = strings.Trim(percent, "%")

	percentf, errs := strconv.ParseFloat(percent, 32)
	if errs != nil {
		return 0.0, errs
	}
	return float32(percentf), nil
}

func Start(config *AppConfigOpt) {

	loginmeta := feixiaohao.UserLoginMeta{
		UserID:     config.UserName,
		PassWD:     config.PassWD,
		IsRemember: false,
	}
	cookies, err := feixiaohao.Login(loginmeta)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// start renew task
	quit := RenewCookies(cookies, loginmeta)

	aliopts := aliyun.AliyunSMSOpt{
		AccessKey:    config.AccessKey,
		AccessID:     config.AccessID,
		SignName:     config.SignName,
		TemplateCode: config.TemplateCode,
		NotifyPhone:  config.NotifyPhone,
	}
	filter := feixiaohao.CoinFilter{
		CoinType:   config.CoinTypes,
		High:       config.PriceHighPercent,
		Low:        config.PriceLowPercent,
		Amplitude:  config.PriceAmplitude,
		TimePeriod: config.NotifyTimePeriod,
	}
	taskctx := &TaskContext{
		LastNotifyTime: make(map[string]int64),
		LastRecord:     make(map[string]float32),
		Cookies:        cookies,
		Filter:         filter,
		AliyunCtx:      aliopts,
	}

	// define quit signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGKILL, syscall.SIGTERM)

	exit := make(chan struct{})
	go func() {
		<-sigs
		quit <- struct{}{}
		exit <- struct{}{}
	}()

	errc := make(chan error, 2)
	timer := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-timer.C:
			Task(taskctx, errc)
		case erri := <-errc:
			fmt.Println("error happend: %s", erri)
		case <-exit:
			return
		}
	}

}

var email string
var author string
var version string

func main() {

	app := cli.NewApp()
	app.Name = "Coin Price Notifier"
	app.Version = version
	app.Author = author
	app.Email = email
	app.Usage = "dynamic detect coin price and notify through sms"
	app.HideHelp = true

	cli.AppHelpTemplate = helpTemplate

	appOptions := &AppConfigOpt{}
	if err := common.ApplyDefaultValues(appOptions); err != nil {
		exit(err, 1)
	}

	cliFlags, flagMappings, err := common.GenerateFlags(appOptions)
	if err != nil {
		exit(err, 3)
	}

	app.Flags = append(
		cliFlags,
		cli.StringFlag{
			Name:   "config",
			Value:  "config.yaml",
			Usage:  "Config file path",
			EnvVar: "COIN_CONFIG",
		},
	)

	app.Action = func(c *cli.Context) {

		configFile := c.String("config")
		_, err := os.Stat(homedir.Expand(configFile))
		if configFile != "config.yaml" || !os.IsNotExist(err) {
			if err := common.ApplyConfigFileYaml(configFile, appOptions); err != nil {
				exit(err, 2)
			}
		}

		common.ApplyFlags(cliFlags, flagMappings, c, appOptions)
		Start(appOptions)
	}

	app.Run(os.Args)
}

func exit(err error, code int) {
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}
