package main

import (
	"log"

	"github.com/fasthttp/router"
	"github.com/masonbrott/sigilkv/handlers"
	"github.com/masonbrott/sigilkv/transaction"
	"github.com/valyala/fasthttp"
)

func main() {
	transaction.InitializeTransactionLog()

	r := router.New()
	r.PUT("/v1/key/{key}", handlers.KVPutHandler)
	r.GET("/v1/key/{key}", handlers.KVGetHandler)
	r.DELETE("/v1/key/{key}", handlers.KVDeleteHandler)

	log.Fatal(fasthttp.ListenAndServe(":8080", r.Handler))
}
