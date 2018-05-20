// check coin price and then notify by send sms
package main

import (
	"os"
	"fmt"
	"time"
	"strings"
	"strconv"
	"syscall"
	"net/http"
	"os/signal"

	"github.com/smileboywtu/CoinNotify/feixiaohao"
	"github.com/smileboywtu/CoinNotify/aliyun"
)

const ACCESSID = "LTAIVi05hIg80A8K"
const ACCESSKEY = "YIGevqWhISrPBqf4nyCO3Wt9DfHVzc"

type TaskContext struct {
	LastNotifyTime int64
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
		if !DoCheck(meta.Percent, ctx.Filter) {
			if (ctx.LastNotifyTime == 0) ||
				(ctx.LastNotifyTime > 0 && time.Now().Unix()-ctx.LastNotifyTime >= ctx.Filter.TimePeriod) {
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

				ctx.LastNotifyTime = time.Now().Unix()
			}
		}
	}
}

func DoCheck(percent string, filter feixiaohao.CoinFilter) bool {
	// strip space
	percent = strings.TrimSpace(percent)

	// trim %
	percent = strings.Trim(percent, "%")

	percentf, errs := strconv.ParseFloat(percent, 32)
	if errs != nil {
		return true
	}

	if float32(percentf) >= filter.High || float32(percentf) <= filter.Low {
		return false
	}

	return true
}

func main() {

	loginmeta := feixiaohao.UserLoginMeta{
		UserID:     "17671601524",
		PassWD:     "chorescb",
		IsRemember: false,
	}
	cookies, err := feixiaohao.Login(loginmeta)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	// start renew task
	quit := RenewCookies(cookies, loginmeta)

	aliopts := aliyun.AliyunSMSOpt{
		AccessKey:    ACCESSKEY,
		AccessID:     ACCESSID,
		SignName:     "房产价格监控",
		TemplateCode: "SMS_135043012",
		NotifyPhone:  "17671601524",
	}
	filter := feixiaohao.CoinFilter{
		[]string{"CMT", "IOST"},
		5.0,
		-2,
		1800,
	}
	taskctx := &TaskContext{
		0,
		cookies,
		filter,
		aliopts,
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
	timer := time.NewTicker(1 * time.Second)
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
