package main

type AppConfigOpt struct {
	// aliyun config
	AccessKey    string `yaml:"accesskey" flagName:"accesskey" flagSName:"ak" flagDescribe:"Aliyun SMS AccessKey" default:""`
	AccessID     string `yaml:"accessid" flagName:"accessid" flagSName:"ai" flagDescribe:"Aliyun SMS AccessID" default:""`
	SignName     string `yaml:"signname" flagName:"signname" flagSName:"sn" flagDescribe:"Aliyun SMS Sign Name" default:""`
	TemplateCode string `yaml:"templatecode" flagName:"templatecode" flagSName:"tc" flagDescribe:"Aliyun SMS Template Code" default:""`

	// feixiaohao
	UserName string `yaml:"userid" flagName:"userid" flagSName:"u" flagDescribe:"Feixiaohao userid" default:""`
	PassWD   string `yaml:"passwd" flagName:"passwd" flagSName:"p" flagDescribe:"Feixiaohao password" default:""`

	// notify

	NotifyPhone      string  `yaml:"notifyphone" flagName:"notifyphone" flagSName:"np" flagDescribe:"User notify phone number" default:""`
	NotifyTimePeriod int64   `yaml:"notifytimeperiod" flagName:"notifytimeperiod" flagSName:"ntp" flagDescribe:"SMS notify time period" default:"3600"`
	PriceLowPercent  float32 `yaml:"lowpricepercent" flagName:"lowpricepercent" flagSName:"lp" flagDescribe:"Coin Price lowest percent" default:"-2.0"`
	PriceHighPercent float32 `yaml:"highpricepercent" flagName:"highpricepercent" flagSName:"hp" flagDescribe:"Coin Price high percent" default:"3.0"`
	PriceAmplitude   float32 `yaml:"amplitude" flagName:"amplitude" flagSName:"apt" flagDescribe:"Coin Price amplitude" default:"1.0"`

	CoinTypes []string `yaml:"cointype" flagName:"cointype" flagSName:"ct" flagDescribe:"Monitor coin type list" default:""`
}
