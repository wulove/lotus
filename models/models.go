package models

import (
	"github.com/astaxie/beego"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/go-xorm/xorm"
	""
)

var x *xorm.Engine

func setEngine() {

}
