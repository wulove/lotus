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

	beego.Informational("AppPath:", beego.AppPath)
	beego.Informational(beego.AppName, setting.APP_VER)

	beego.InsertFilter("/capata/*", beego.BeforeRouter, setting.Captcha.Handler)

	beego.Run()
}
