package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/namsral/flag"
	"github.com/rs/zerolog/log"
	kak "github.com/segmentio/kafka-go"
	"github.com/yusufsyaifudin/go-kafka-example/dep/kafka"
)

var db *gorm.DB
var messages []string

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
	order struct {
		gorm.Model
		//OrderID uint               `json:"id"`
		Product      product `gorm:"foreignkey:productRefer`
		Order_Amount int     `json:"order_amount"`
	}
	customer struct {
		Name    string
		Address string
		PinCode string
	}
	message struct {
		OrderID         string
		ProductID       string
		ProductName     string
		ProductQuantity string
		TotalAmount     string
		CustomerName    string
		CustomerAddress string
		CustomerPinCode string
	}
)

func init() {
	var err error

	db, err = gorm.Open("mysql", "root:Shon@2544@/ordermanagement?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("failed to connect to database")
	}
	db.AutoMigrate(&product{})
	db.AutoMigrate(&order{})
}

//kafka variables
var logger = log.With().Str("pkg", "main").Logger()

var (
	listenAddrApi string

	// kafka
	kafkaBrokerUrl string
	kafkaVerbose   bool
	kafkaClientId  string
	kafkaTopic     string
)

// kafka init

func main() {
	// kafka init
	flag.StringVar(&listenAddrApi, "listen-address", "0.0.0.0:9000", "Listen address for api")
	flag.StringVar(&kafkaBrokerUrl, "kafka-brokers", "localhost:9092,localhost:9093,localhost:9094,localhost:9095", "Kafka brokers in comma separated value")
	flag.BoolVar(&kafkaVerbose, "kafka-verbose", true, "Kafka verbose logging")
	flag.StringVar(&kafkaClientId, "kafka-client-id", "my-kafka-client", "Kafka client id to connect")
	flag.StringVar(&kafkaTopic, "kafka-topic", "foo", "Kafka topic to push")

	flag.Parse()

	// connect to kafka
	kafkaProducer, err := kafka.Configure(strings.Split(kafkaBrokerUrl, ","), kafkaClientId, kafkaTopic)
	if err != nil {
		logger.Error().Str("error", err.Error()).Msg("unable to configure kafka")
		return
	}
	defer kafkaProducer.Close()

	router := gin.Default()
	router.Use(static.Serve("/", static.LocalFile("./view", true)))

	v1 := router.Group("/api/v1/products")
	v2 := router.Group("/api/v2/orders")

	v1.POST("/", addProduct)
	v1.GET("/", getProducts)
	v2.POST("/", placeOrder)
	router.GET("/shipping", getShippingOrder)

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
		_products = append(_products, transformedProduct{ProductName: item.ProductName, ProductQuantity: item.ProductQuantity, ProductPrice: item.ProductPrice})
	}
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"data":   _products,
	})
}

func placeOrder(c *gin.Context) {
	var message message
	var cust_info customer

	cust_info.Name = "shon"
	cust_info.Address = "Bellandur"
	cust_info.PinCode = "560103"

	pname := c.PostForm("pname")
	//fmt.Println("............asdsf............", pname)
	//pname := "Mirinda"
	// := 5
	qty, _ := strconv.Atoi(c.PostForm("pqty"))

	avail, prod, amt := isProductAvailable(pname, qty)
	message.OrderID = "order-id"
	message.ProductID = "product-id"
	message.ProductName = string(prod.ProductName)
	message.ProductQuantity = strconv.Itoa(qty)
	message.TotalAmount = strconv.Itoa(amt)
	message.CustomerName = cust_info.Name
	message.CustomerAddress = cust_info.Address
	message.CustomerPinCode = cust_info.PinCode

	if avail {

		fmt.Println("placed order is succefull...")
		ord := order{

			Product:      prod,
			Order_Amount: amt,
		}
		fmt.Println("order for table : ", ord)
		db.Save(&ord)

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "placed order is succeful...!",
			"data":    ord,
		})

		//posting message to kafka

		postDataToKafka(c, message)
		return
	} else {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No Product found !!"})
		return
	}

}
func isProductAvailable(pname string, qty int) (bool, product, int) {
	available := false
	amt := 0
	var prod product
	db.Where("product_name = ?", pname).First(&prod)

	if pname == prod.ProductName && qty <= prod.ProductQuantity {
		available = true
		amt = prod.ProductPrice * qty
		remainingQuantity := prod.ProductQuantity - qty
		db.Model(&prod).Update("product_quantity", remainingQuantity)
		return available, prod, amt
	}
	return available, prod, amt
}

func postDataToKafka(ctx *gin.Context, message message) {
	parent := context.Background()
	defer parent.Done()

	ctx.Bind(message)
	formInBytes, err := json.Marshal(message)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": map[string]interface{}{
				"message": fmt.Sprintf("error while marshalling json: %s", err.Error()),
			},
		})

		ctx.Abort()
		return
	}

	err = kafka.Push(parent, nil, formInBytes)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": map[string]interface{}{
				"message": fmt.Sprintf("error while push message into kafka: %s", err.Error()),
			},
		})

		ctx.Abort()
		return
	}

}

func getShippingOrder(c *gin.Context) {

	//var (
	// 	// kafka
	// 	kafkaBrokerUrl     string
	// 	kafkaVerbose       bool
	// 	kafkaTopic         string
	//	kafkaConsumerGroup string
	// 	kafkaClientId      string
	// )

	// flag.StringVar(&kafkaBrokerUrl, "kafka-brokers", "localhost:19092,localhost:29092,localhost:39092", "Kafka brokers in comma separated value")
	// flag.BoolVar(&kafkaVerbose, "kafka-verbose", true, "Kafka verbose logging")
	// flag.StringVar(&kafkaTopic, "kafka-topic", "foo", "Kafka topic. Only one topic per worker.")
	//flag.StringVar(&kafkaConsumerGroup, "kafka-consumer-group", "consumer-group", "Kafka consumer group")
	// flag.StringVar(&kafkaClientId, "kafka-client-id", "my-client-id", "Kafka client id")

	//flag.Parse()

	brokers := strings.Split(kafkaBrokerUrl, ",")
	config := kak.ReaderConfig{
		Brokers:         brokers,
		GroupID:         kafkaClientId,
		Topic:           kafkaTopic,
		MinBytes:        10e3,            // 10KB
		MaxBytes:        10e6,            // 10MB
		MaxWait:         1 * time.Second, // Maximum amount of time to wait for new data to come when fetching batches of messages from kafka.
		ReadLagInterval: -1,
	}

	reader := kak.NewReader(config)
	defer reader.Close()

	//for {
	m, err := reader.ReadMessage(context.Background())
	if err != nil {
		log.Error().Msgf("error while receiving message: %s", err.Error())
		//		continue
		return
	}

	value := m.Value
	// if m.CompressionCodec == snappy.NewCompressionCodec() {
	// 	_, err = snappy.NewCompressionCodec().Decode(value, m.Value)
	// }

	// if err != nil {
	// 	log.Error().Msgf("error while receiving message: %s", err.Error())
	// 	continue
	// }

	fmt.Printf("message at topic/partition/offset %v/%v/%v: %s\n", m.Topic, m.Partition, m.Offset, string(value))
	messages = append(messages, string(value))
	//}

	c.JSON(200, messages)

}
