package order

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	products "github.com/mkhalegaonkar3/product-service-go/products"
)

type order struct {
	gorm.Model
	Product     products.TransformedProduct `gorm:"foreignkey:productRefer`
	OrderAmount int                         `json:"order_amount"`
}

// PlaceOrder func
func PlaceOrder(c *gin.Context) {
	pname := c.PostForm("pname")
	//fmt.Println("............asdsf............", pname)
	//pname := "Mirinda"
	// := 5
	qty, _ := strconv.Atoi(c.PostForm("pqty"))

	avail, prod, amt := products.IsProductAvailable(pname, qty)
	if avail {

		fmt.Println("placed order is succefull...")
		ord := order{

			Product: products.TransformedProduct{
				ProductID:       prod.ID,
				ProductName:     pname,
				ProductQuantity: qty,
				ProductPrice:    prod.ProductPrice,
			},
			OrderAmount: amt,
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "placed order is succeful...!",
			"data":    ord,
		})

	} else {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No Product found !!"})
		return
	}

}
