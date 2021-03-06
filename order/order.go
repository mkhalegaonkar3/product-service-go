package order

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/mkhalegaonkar3/product-service-go/kafkaconfig"
	model "github.com/mkhalegaonkar3/product-service-go/model"
	products "github.com/mkhalegaonkar3/product-service-go/products"
	"github.com/rs/zerolog/log"
)

type order struct {
	gorm.Model
	Product     products.TransformedProduct `gorm:"foreignkey:productRefer`
	OrderAmount int                         `json:"order_amount"`
}

var (
	logger = log.With().Str("pkg", "main").Logger()
)

type customer struct {
	Name    string
	Address string
	PinCode string
}

var custInfo = customer{
	Name:    "shon",
	Address: "Bellandur",
	PinCode: "560103",
}

// PlaceOrder func
func PlaceOrder(c *gin.Context) {
	pname := c.PostForm("pname")
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
		model.Db.Save(&ord)
		kafkaProducer, err := kafkaconfig.Configure(strings.Split(kafkaconfig.KafkaBrokerUrl, ","), kafkaconfig.KafkaClientId, kafkaconfig.KafkaTopic)
		if err != nil {
			logger.Error().Str("error", err.Error()).Msg("unable to configure kafka")
			return
		}
		defer kafkaProducer.Close()

		msg := kafkaconfig.Message{
			OrderID:         ord.ID,
			ProductID:       ord.Product.ProductID,
			ProductName:     ord.Product.ProductName,
			ProductQuantity: ord.Product.ProductQuantity,
			TotalAmount:     ord.OrderAmount,
			CustomerName:    custInfo.Name,
			CustomerAddress: custInfo.Address,
			CustomerPinCode: custInfo.PinCode,
		}
		kafkaconfig.PostDataToKafka(c, msg)
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "placed order is successfull...!",
			"data":    ord,
		})

	} else {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No Product found !!"})
		return
	}

}
