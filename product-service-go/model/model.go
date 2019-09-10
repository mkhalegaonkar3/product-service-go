package model

import (
	"github.com/jinzhu/gorm"
	//TODO
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var Db *gorm.DB

func init() {
	var err error

	Db, err = gorm.Open("mysql", "root:Shon@2544@/ordermanagement?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("failed to connect to database")
	}
	
}
