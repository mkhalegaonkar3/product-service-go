package model

import (
	"github.com/jinzhu/gorm"
	//TODO
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

//creating DB var
var Db *gorm.DB

func init() {
	var err error
	
	Db, err = gorm.Open("mysql", "root:Shon@2544@/OrrderManagement?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("failed to connect to database")
	}
	Db.AutoMigrate(&products.product{})
	Db.AutoMigrate(&order.order{})
}
