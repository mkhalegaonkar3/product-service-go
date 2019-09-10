FROM golang:latest

# Copy the local package files to the container’s workspace.

WORKDIR /go/src/github.com/mkhalegaonkar3
RUN cd /go/src/github.com/mkhalegaonkar3 \
    && git https://github.com/mkhalegaonkar3/product-service-go.git

RUN cd /go/src/github.com/mkhalegaonkar3/product-service-go
# Install our dependencies
RUN go get github.com/go-sql-driver/mysql  
RUN go get github.com/gin-gonic/gin
RUN go get github.com/segmentio/kafka-go
RUN go get github.com/segmentio/kafka-go/snappy
RUN go get github.com/jinzhu/gorm/dialects/mysql
RUN go get github.com/jinzhu/gorm
RUN go get github.com/rs/zerolog/log
RUN go get github.com/gin-gonic/contrib/static
RUN go get github.com/zsais/go-gin-prometheus

# Install api binary globally within container 
RUN go install github.com/mkhalegaonkar3/product-service-go

# Set binary as entrypoint
ENTRYPOINT /go/bin/product-service-go

# Expose default port (8888)
EXPOSE 8888 