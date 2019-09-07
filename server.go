package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

type product struct {
	gorm.Model
	ProductName     string `json:"product_name"`
	ProductQuantity int    `json:"product_quantity"`
	ProductPrice    int    `json:"product_price"`
}

func init() {
	var err error

	db, err = gorm.Open("mysql", "root:root@/OrderManagement?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("failed to connect to database")
	}
	db.AutoMigrate(&product{})
}

func main() {

	router := gin.Default()

	v1 := router.Group("/api/v1/products")

	v1.POST("/", addProduct)
	// v1.GET("/", getProducts)

	router.Run()
}
func addProduct(c *gin.Context) {
	price, _ := strconv.Atoi(c.PostForm("pprice"))
	quantity, _ := strconv.Atoi(c.PostForm("pquantity"))
	prod := product{

		ProductName:     c.PostForm("pname"),
		ProductPrice:    price,
		ProductQuantity: quantity,
	}
	db.Save(&prod)
	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "Product added successfully",
	})

}

// func getProducts(c *gin.Context) {

// }
