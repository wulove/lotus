package main

import (
	"github.com/astaxie/beego"
	_ "github.com/wulove/lotus/routers"
	"github.com/wulove/lotus/setting"
)

func initialize() {
	setting.LoadConfig()
}

func main() {

	initialize()

	beego.Info("AppPath:", beego.AppPath)
	beego.Info(beego.AppName, setting.APP_VER)

	beego.InsertFilter("/capata/*", beego.BeforeRouter, setting.Captcha.Handler)

	beego.Run()
}
