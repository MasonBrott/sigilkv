package handlers

import (
	"errors"

	"github.com/masonbrott/sigilkv/core"
	"github.com/masonbrott/sigilkv/transaction"
	"github.com/valyala/fasthttp"
)

func KVPutHandler(ctx *fasthttp.RequestCtx) {
	key := ctx.UserValue("key").(string)
	value := string(ctx.PostBody())

	err := core.Put(key, string(value))
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	transaction.TranLogger.WritePut(key, value)

	ctx.SetStatusCode(201)
}

func KVGetHandler(ctx *fasthttp.RequestCtx) {
	key := ctx.UserValue("key").(string)

	value, err := core.Get(key)
	if errors.Is(err, core.ErrorNoSuchKey) {
		ctx.Error(err.Error(), 404)
		return
	}

	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	ctx.Write([]byte(value))
}

func KVDeleteHandler(ctx *fasthttp.RequestCtx) {
	key := ctx.UserValue("key").(string)

	err := core.Delete(key)
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	transaction.TranLogger.WriteDelete(key)

	ctx.SetStatusCode(200)
}
