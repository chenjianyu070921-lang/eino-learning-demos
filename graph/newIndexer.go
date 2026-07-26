package main

import (
	"context"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino-ext/components/indexer/milvus2"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

// IndexerClient 创建 milvus2 索引器。embedder 使用 embedding.Embedder 接口，
// 因此无论是 ollama 还是 ark 都能传入。Dimension 必须与 embedding 模型输出维度一致。
func IndexerClient(ctx context.Context, newClient *milvusclient.Client, embedder embedding.Embedder) *milvus2.Indexer {
	indexerConfig := milvus2.IndexerConfig{
		Client:     newClient, //Milvus 客户端实例
		Collection: "test",    //目标集合名称
		Embedding:  embedder,  //Embedding 向量化实例
		Vector: &milvus2.VectorConfig{
			Dimension:    1024,                                                          // 与 bge-m3 输出维度匹配
			MetricType:   milvus2.COSINE,                                                //距离度量方式，用于向量相似度检索
			IndexBuilder: milvus2.NewHNSWIndexBuilder().WithM(16).WithEfConstruction(200), //用来定义 Milvus 向量索引类型，你这里使用 HNSW（主流高性能内存索引）
		},
	}

	indexer, err := milvus2.NewIndexer(ctx, &indexerConfig)
	if err != nil {
		panic(err)
	}
	return indexer
}
