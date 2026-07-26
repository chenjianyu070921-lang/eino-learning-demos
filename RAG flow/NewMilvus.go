package main

import (
	"context"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

func MilvusClient(ctx context.Context) *milvusclient.Client {
	newClient, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: "127.0.0.1:19530",
		DBName:  "default",
	})
	if err != nil {
		panic(err)
	}
	return newClient
}
