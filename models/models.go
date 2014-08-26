package models

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

func CountObjects(qs orm.QuerySeter) (int64, error) {
	n, err := qs.Count()
	if err != nil {
		beego.Error("models.CountObjects ", err)
		return 0, err
	}
	return n, nil
}
