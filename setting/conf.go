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
	"github.com/astaxie/beego/utils/captcha"
	"github.com/howeyc/fsnotify"
)

const (
	APP_VER = "0.0.1"
)

var (
	IsProMode bool
	AppName   string
	AppLogo   string
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

	beego.RunMode = Cfg.MustValue("app", "run_mode")
	beego.HttpPort = Cfg.MustInt("app", "http_port")

	IsProMode = beego.RunMode == "pro"
	if IsProMode {
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
		beego.SessionSavePath == Cfg.MustValue("session", "session_path", "sessions")
	}
	beego.SessionName = Cfg.MustValue("session", "session_name", "lotus_sess")
	beego.SessionCookieLifeTime = Cfg.MustInt("session", "session_life_time", 0)
	beego.SessionGCMaxLifetime = Cfg.MustInt64("session", "session_gc_time", 86400)

	beego.EnableXSRF = Cfg.MustBool("xsrf", "xsrf_on")
	if beego.EnableXSRF {
		beego.XSRFKEY = Cfg.MustValue("xsrf", "xsrf_key", "lotus_wulove")
		beego.XSRFExpire = Cfg.MustInt64("xsrf", "xsrf_expire", 86400*30)
	}

	reloadConfig()

	configWatcher()

}

func reloadConfig() {
	AppName = Cfg.MustValue("app", "app_name", "Lotus")
	beego.AppName = AppName

	AppLogo = Cfg.MustValue("app", "app_logo")
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
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		beego.Error("Failed to get file information[ %s ]\n", err)
		return nil, err
	}

	return fi.ModTime().Unix()

}

func setLogs() {
	log := logs.NewLogger(10000)
	logconf := new(map[string]interface{})
	logconf["filename"] = "logs/log.log"
	logconf["level"] = logs.LevelInfo
	logconf["maxsize"] = 1 << 35
	logconf["maxdays"] = 30

	config, _ := json.Marshal(logconf)

	log.SetLogger("file", string(config))
	beego.BeeLogger = log
	beego.SetLogFuncCall(true)
}
