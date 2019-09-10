package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/mkhalegaonkar3/product-service-go/kafkaconfig"
	order "github.com/mkhalegaonkar3/product-service-go/order"
	products "github.com/mkhalegaonkar3/product-service-go/products"
	shipping "github.com/mkhalegaonkar3/product-service-go/shipping"
	"github.com/namsral/flag"
	"github.com/rs/zerolog/log"
)

var (
	//listenAddrApi string

	// kafka
	KafkaBrokerUrl string
	kafkaVerbose   bool
	KafkaClientId  string
	KafkaTopic     string
	Messages       []string
)

// used to match which service is being called
const (
	ADDPRODUCT  = "/addProduct"
	GETPRODUCTS = "/getProducts"
	PLACEORDER  = "/placeOrder"
	SHIPPING    = "/getShipping"
)

func main() {
	var logger = log.With().Str("pkg", "main").Logger()
	//flag.StringVar(&listenAddrApi, "listen-address", "0.0.0.0:9000", "Listen address for api")
	flag.StringVar(&KafkaBrokerUrl, "kafka-brokers", "localhost:9092,localhost:9093,localhost:9094,localhost:9095", "Kafka brokers in comma separated value")
	flag.BoolVar(&kafkaVerbose, "kafka-verbose", true, "Kafka verbose logging")
	flag.StringVar(&KafkaClientId, "kafka-client-id", "my-kafka-client", "Kafka client id to connect")
	flag.StringVar(&KafkaTopic, "kafka-topic", "foo", "Kafka topic to push")
	flag.Parse()

	kafkaProducer, err := kafkaconfig.Configure(strings.Split(KafkaBrokerUrl, ","), KafkaClientId, KafkaTopic)
	if err != nil {
		logger.Error().Str("error", err.Error()).Msg("unable to configure kafka")
		return
	}
	defer kafkaProducer.Close()
	rand.Seed(time.Now().UTC().UnixNano())
	router := initRouter()
	router.Run(":8000")
}

func initRouter() *gin.Engine {

	r := gin.New()
	r.Use(static.Serve("/", static.LocalFile("./view", true)))
	r.Use(gin.Recovery(), plainLoggerWithWriter(gin.DefaultWriter))

	r.GET("/status", statusCheck)
	r.POST("/addProduct", requestRouter)
	r.GET("/getProducts", requestRouter)
	r.POST("/placeOrder", requestRouter)

	return r
}

// PlainLoggerWithWriter mimics the Gin LoggerWithWriter without the colors
func plainLoggerWithWriter(out io.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		if c.Request.URL.Path != "/status" {
			fmt.Fprintf(out, "%s [%s] %s [%v] \"%s %s %s\" %d %d %v %s %s %s \"%s\"\n",
				c.ClientIP(),
				c.Request.UserAgent(),
				c.Request.Header.Get(gin.AuthUserKey),
				end.Format("02/Jan/2006:15:04:05 -0700"),
				c.Request.Method,
				c.Request.URL.Path,
				c.Request.Proto,
				c.Writer.Status(),
				c.Writer.Size(),
				fmt.Sprintf("%.4f", latency.Seconds()),
				c.Request.Header.Get("RequestType"),
				c.Request.Header.Get("ResponseSource"),
				c.Request.Form.Encode(),
				c.Request.Header.Get("ResponseBody"),
			)
		}
	}
}

// statusCheck returns a 200/OK when called if we can contact the be env
func statusCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func exception(c *gin.Context) {
	c.JSON(500, gin.H{"success": false, "error": "Unable to process order"})
}

func requestRouter(c *gin.Context) {

	path := c.Request.URL.Path
	fmt.Println("The obtained path is:- ", path)
	switch path {
	case ADDPRODUCT:
		products.AddProduct(c)
	case GETPRODUCTS:
		products.GetProducts(c)
	case PLACEORDER:
		order.PlaceOrder(c)
	case SHIPPING:
		shipping.GetShippingDetails(c)
	}
}
