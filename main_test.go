package main

import (
	"testing"
	"github.com/smileboywtu/CoinNotify/feixiaohao"
	"github.com/smileboywtu/CoinNotify/aliyun"
	"time"
)

func TestNeedNotify(t *testing.T) {

	// first time need to notify
	ctx := TaskContext{
		LastNotifyTime:make(map[string]int64),
		LastRecord: make(map[string]float32),
		Cookies:nil,
		Filter:feixiaohao.CoinFilter{
			TimePeriod:2,
		},
		AliyunCtx: aliyun.AliyunSMSOpt{
		},
	}

	meta:= feixiaohao.CoinPriceMeta{
		Price: "3.2",
		Percent: "5.2%",
		CoinType: "CMT",
		Platform: "Bettrix",
	}

	notify, pricef := NeedNotify(meta, ctx)
	if pricef == 0 || !notify{
		t.Fatal("first time notify error")
	}

	ctx.LastNotifyTime[meta.CoinType] = time.Now().Unix()
	ctx.LastRecord[meta.CoinType] = pricef
	t.Log("first time notify", pricef)

	// second if percent larger than threhold
	meta.Percent = "7.2%"
	notify, pricef = NeedNotify(meta, ctx)
	if pricef == 0 || !notify{
		t.Fatal("amplitude larger test error, task context: ", ctx)
	}

	ctx.LastNotifyTime[meta.CoinType] = time.Now().Unix()
	ctx.LastRecord[meta.CoinType] = pricef
	t.Log("amplitude notify", pricef)


	// third if percent lower than threhold
	meta.Percent = "3.2%"
	notify, pricef = NeedNotify(meta, ctx)
	if pricef == 0 || !notify{
		t.Fatal("amplitude lower test error, task context: ", ctx)
	}

	ctx.LastNotifyTime[meta.CoinType] = time.Now().Unix()
	ctx.LastRecord[meta.CoinType] = pricef
	t.Log("amplitude notify", pricef)


	// time wait
	done := make(chan bool)
	time.AfterFunc(3, func() {

		notify, pricef = NeedNotify(meta, ctx)
		if pricef == 0 || !notify{
			t.Fatal("time threhold test error, task context: ", ctx)
		}

		done <- true
	})
	<- done
	t.Log("timeout notify", pricef)

}