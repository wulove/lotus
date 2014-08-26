package setting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Unknwon/goconfig"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils/captcha"
	"github.com/beego/i18n"
	"github.com/howeyc/fsnotify"

	_ "github.com/go-sql-driver/mysql"
)

const (
	APP_VER = "0.0.1"
)

var (
	IsProMode bool
	AppName   string
	AppLogo   string
	AppUrl    string
	Langs     []string

	LoginRememberDays  int
	CookieRemerberName string
	CookieUserName     string
	LoginMaxRetries    int
)

var (
	Cfg     *goconfig.ConfigFile
	Cache   cache.Cache
	Captcha *captcha.Captcha
)

var (
	AppConfPath = "conf/app.ini"
)

func LoadConfig() {
	var err error
	Cfg, err = goconfig.LoadConfigFile(AppConfPath)
	if err != nil {
		fmt.Println("Fail to load Configuration file: " + err.Error())
		os.Exit(2)
	}

	Cfg.BlockMode = false

	beego.BeegoServerName = "lotus:" + APP_VER
	beego.RunMode = Cfg.MustValue("app", "run_mode", "dev")
	beego.HttpPort = Cfg.MustInt("app", "http_port", 8080)

	IsProMode = beego.RunMode == "pro"
	if IsProMode {
		beego.EnableGzip = true
		setLogs()
	}
	// cache system
	Cache, err = cache.NewCache("memory", `{"interval":360}`)

	Captcha = captcha.NewCaptcha("/captcha/", Cache)
	Captcha.FieldIdName = "CaptchaId"
	Captcha.FieldCaptchaName = "Captcha"

	beego.SessionOn = true
	beego.SessionProvider = Cfg.MustValue("session", "session_provider", "memory")
	if beego.SessionProvider == "file" || beego.SessionProvider == "mysql" || beego.SessionProvider == "redis" {
		beego.SessionSavePath = Cfg.MustValue("session", "session_path", "sessions")
	}
	beego.SessionName = Cfg.MustValue("session", "session_name", "lotus_sess")
	beego.SessionCookieLifeTime = Cfg.MustInt("session", "session_life_time", 0)
	beego.SessionGCMaxLifetime = Cfg.MustInt64("session", "session_gc_time", 86400)

	beego.EnableXSRF = Cfg.MustBool("xsrf", "xsrf_on", true)
	if beego.EnableXSRF {
		beego.XSRFKEY = Cfg.MustValue("xsrf", "xsrf_key", "lotus_wulove")
		beego.XSRFExpire = Cfg.MustInt("xsrf", "xsrf_expire", 86400*30)
	}

	driverName := Cfg.MustValue("orm", "driver_name", "mysql")
	dataSource := Cfg.MustValue("orm", "data_source", "root:root@/lotus?charset=utf8&loc=UTC")
	maxIdle := Cfg.MustInt("orm", "max_idle_conn", 30)
	maxOpen := Cfg.MustInt("orm", "max_open_conn", 50)

	err = orm.RegisterDataBase("default", driverName, dataSource, maxIdle, maxOpen)
	if err != nil {
		beego.Error(err)
	}
	orm.RunCommand()
	err = orm.RunSyncdb("default", false, false)
	if err != nil {
		beego.Error(err)
	}
	if !IsProMode {
		orm.Debug = true
	}

	reloadConfig()

	configWatcher()

}

func reloadConfig() {
	AppName = Cfg.MustValue("app", "app_name", "Lotus")
	beego.AppName = AppName
	AppLogo = Cfg.MustValue("app", "app_logo")
	AppUrl = Cfg.MustValue("app", "app_url")

	LoginMaxRetries = Cfg.MustInt("app", "login_max_retries", 3)

	LoginRememberDays = Cfg.MustInt("app", "login_remember_days", 7)
	CookieRemerberName = Cfg.MustValue("app", "cookie_remember_name", "lotus_magic")
	CookieUserName = Cfg.MustValue("app", "cookie_user_name", "lotus_power")
}

func settingLocales() {
	// load locales with locale_LANG.ini files
	langs := "en-US|zh-CN"
	for _, lang := range strings.Split(langs, "|") {
		lang = strings.TrimSpace(lang)
		files := []string{"conf/" + "locale_" + lang + ".ini"}
		if fh, err := os.Open(files[0]); err == nil {
			fh.Close()
		} else {
			files = nil
		}
		if err := i18n.SetMessage(lang, "conf/global/"+"locale_"+lang+".ini", files...); err != nil {
			beego.Error("Fail to set message file: " + err.Error())
			os.Exit(2)
		}
	}
	Langs = i18n.ListLangs()
}

var eventTime = make(map[string]int64)

func configWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic("Failed start app watcher: " + err.Error())
	}

	go func() {
		for {
			select {
			case event := <-watcher.Event:
				switch filepath.Ext(event.Name) {
				case ".ini":
					if checkEventTime(event.Name) {
						continue
					}
					beego.Info(event)
					if err := Cfg.Reload(); err != nil {
						beego.Error("Conf Reload: ", err)
					}

					reloadConfig()
					beego.Info("Config Reloaded!")
				}
			}
		}
	}()

	if err := watcher.AddWatch("conf", fsnotify.FSN_MODIFY); err != nil {
		beego.Error("Watch dirpath(conf): ", err)
	}
}

// checkEventTime return true if FileModTime does not change
func checkEventTime(name string) bool {
	mt, err := getFileModTime(name)
	if err != nil {
		return true
	}
	if eventTime[name] == mt {
		return true
	}

	eventTime[name] = mt
	return false
}

func getFileModTime(path string) (int64, error) {
	path = strings.Replace(path, "\\", "/", -1)
	f, err := os.Open(path)
	if err != nil {
		beego.Error("Failed to open file[ %s ]\n", err)
		return 0, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		beego.Error("Failed to get file information[ %s ]\n", err)
		return 0, err
	}

	return fi.ModTime().Unix(), nil

}

func setLogs() {
	log := logs.NewLogger(10000)
	logconf := make(map[string]interface{})
	logconf["filename"] = "logs/log.log"
	logconf["level"] = logs.LevelInformational
	logconf["maxsize"] = 32 << 20
	logconf["maxdays"] = 30

	config, _ := json.Marshal(logconf)

	log.SetLogger("file", string(config))
	beego.BeeLogger = log
	beego.SetLogFuncCall(true)
}
