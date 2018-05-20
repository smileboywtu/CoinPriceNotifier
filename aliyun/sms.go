package aliyun

import (
	"fmt"
	"errors"
	"encoding/json"

	"github.com/GiterLab/aliyun-sms-go-sdk/dysms"
	"github.com/tobyzxj/uuid"
)

type AliyunSMSOpt struct {
	AccessKey string
	AccessID  string

	SignName     string
	TemplateCode string
	NotifyPhone  string
}

type SMSContentCtx struct {
	Platform string `json:"platform"`
	CoinType string `json:"cointype"`
	Price    string `json:"price"`
	Percent  string `json:"percent"`
}

func SendSMS(opts AliyunSMSOpt, context SMSContentCtx) error {

	dysms.HTTPDebugEnable = true
	dysms.SetACLClient(opts.AccessID, opts.AccessKey)

	params, err := json.Marshal(context)
	//respSendSms, err := dysms.SendSms(uuid.New(), phone, "", "SMS_135043012", string(params)).DoActionWithException()
	respSendSms, err := dysms.SendSms(uuid.New(), opts.NotifyPhone, opts.SignName, opts.TemplateCode, string(params)).DoActionWithException()
	if err != nil {
		return errors.New(fmt.Sprintf("send sms failed", err, respSendSms.Error()))
	}
	return nil
}
