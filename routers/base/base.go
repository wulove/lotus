package base

import (
	"html/template"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/beego/i18n"
	"github.com/wulove/lotus/models"
	"github.com/wulove/lotus/modules/utils"
	"github.com/wulove/lotus/setting"
)

type NestPreparer interface {
	NestPrepare()
}

type BaseRouter struct {
	beego.Controller
	i18n.Locale
	User    models.User
	IsLogin bool
}

func (this *BaseRouter) Prepare() {

	this.Data["PageStartTime"] = time.Now()
	this.Data["AppName"] = setting.AppName
	this.Data["AppVer"] = setting.AppVer
	this.Data["AppLogo"] = setting.AppLogo
	this.Data["AppUrl"] = setting.AppUrl
	this.Data["IsProMode"] = setting.IsProMode

	this.Data["xsrf_token"] = this.XsrfToken()
	this.Data["xsrf_html"] = template.HTML(this.XsrfFormHtml())

	this.Data["delete_method"] = template.HTML(`<input type="hidden" name="_method" value="DELETE">`)
	this.Data["put_method"] = template.HTML(`<input type="hidden" name="_method" value="PUT">`)

	// Redirect to make URL clean.
	if this.setLang() {
		i := strings.Index(this.Ctx.Request.RequestURI, "?")
		this.Redirect(this.Ctx.Request.RequestURI[:i], 302)
		return
	}

	if this.Ctx.Request.Method == "GET" {
		this.FormOnceCreate()
	}

	if app, ok := this.AppController.(NestPreparer); ok {
		app.NestPrepare()
	}

}

// check form once, void re-submit
func (this *BaseRouter) FormOnceNotMatch() bool {
	notMatch := false
	recreat := false

	var value string
	if vals, ok := this.Input()["_once"]; ok && len(vals) > 0 {
		value = vals[0]
	} else {
		value = this.Ctx.Input.Header("X-Form-Once")
	}

	if v, ok := this.GetSession("form_once").(string); ok && len(v) != 0 {
		if value != v {
			notMatch = true
		} else {
			// if matched then re-creat once
			recreat = true
		}
	}

	this.FormOnceCreate(recreat)
	return notMatch
}

// create form once html
func (this *BaseRouter) FormOnceCreate(args ...bool) {
	var value string
	var creat bool
	creat = len(args) > 0 && args[0]
	if !creat {
		if v, ok := this.GetSession("form_once").(string); ok && len(v) != 0 {
			value = v
		} else {
			creat = true
		}
	}
	if creat {
		value = utils.GetRandomString(10)
		this.SetSession("form_once", value)
	}
	this.Data["once_token"] = value
	this.Data["once_html"] = template.HTML(`<input type="hidden" name="_once" value="` + value + `">`)
}

func (this *BaseRouter) setLangCookie(lang string) {
	this.Ctx.SetCookie("lang", lang, 60*60*24*365, "/", nil, nil, false)
}

func (this *BaseRouter) setlang() bool {
	isNeedRedir := false
	hasCookie := false

	langs := setting.Langs

	// 1. Check URL arguments.
	lang := this.GetString("lang")

	// 2. Get language information from cookies.
	if len(lang) == 0 {
		lang = this.Ctx.GetCookie("lang")
		hasCookie = true
	} else {
		isNeedRedir = true
	}

	// Check again in case someone modify by purpose.
	if !i18n.IsExist(lang) {
		lang = ""
		isNeedRedir = false
		hasCookie = false
	}

	// 3. check if isLogin then use user setting
	if len(lang) == 0 && this.IsLogin {
		// it's need to write after user manager's coding
	}

	// 4. Get language information from 'Accept-Language'.
	if len(lang) == 0 {
		al := this.Ctx.Input.Header("Accept-Language")
		if len(al) > 4 {
			al = al[:5]
			if i18n.IsExist(al) {
				lang = al
			}
		}
	}

	// 5. DefaucurLang language is English.
	if len(lang) == 0 {
		lang = "en-US"
		isNeedRedir = false
	}

	// Save language information in cookies.
	if !hasCookie {
		this.setLangCookie(lang)
	}

	// Set language properties.
	this.Data["Lang"] = lang
	this.Data["Langs"] = langs

	this.Lang = lang

	return isNeedRedir
}
