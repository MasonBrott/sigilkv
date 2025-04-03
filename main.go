package main

import (
	"log"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func main() {
	r := router.New()
	r.PUT("/v1/key/{key}", kvputhandler)
	r.GET("/v1/key/{key}", kvgethandler)
	r.DELETE("/v1/key/{key}", kvdeletehandler)

	log.Fatal(fasthttp.ListenAndServe(":8080", r.Handler))
}
