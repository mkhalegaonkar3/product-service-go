package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

type (
	product struct {
		gorm.Model
		ProductName     string `json:"product_name"`
		ProductQuantity int    `json:"product_quantity"`
		ProductPrice    int    `json:"product_price"`
	}
	transformedProduct struct {
		ProductID       uint   `json:"id"`
		ProductName     string `json:"product_name"`
		ProductQuantity int    `json:"product_quantity"`
		ProductPrice    int    `json:"product_price"`
	}
)

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
	router.Use(static.Serve("/", static.LocalFile("./view", true)))

	v1 := router.Group("/api/v1/products")

	v1.POST("/", addProduct)
	v1.GET("/", getProducts)

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

func getProducts(c *gin.Context) {
	var products []product
	var _products []transformedProduct

	db.Find(&products)
	if len(products) <= 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No products found !!"})
		return
	}
	for _, item := range products {
		_products = append(_products, transformedProduct{ProductID: item.ID, ProductName: item.ProductName, ProductQuantity: item.ProductQuantity, ProductPrice: item.ProductPrice})
	}
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"data":   _products,
	})
}
