package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino-ext/components/embedding/ollama"
	"github.com/joho/godotenv"

	"github.com/cloudwego/eino-ext/components/indexer/milvus2"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

func main() {

	fmt.Println("程序开始")

	err := godotenv.Load()
	if err != nil {
		panic("加载.env文件失败：" + err.Error())
	}

	fmt.Println(".env加载成功")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)

	defer cancel()

	fmt.Println("准备连接Milvus")

	newClient, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: "127.0.0.1:19530",
		DBName:  "default",
	})

	if err != nil {
		panic(err)
	}

	fmt.Println("Milvus连接成功")

	fmt.Println("准备创建Embedding")
	embedder, err := ollama.NewEmbedder(ctx, &ollama.EmbeddingConfig{
		BaseURL: "http://localhost:11434",
		Model:   "bge-m3",
		Timeout: 10 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Embedding创建成功")

	indexerConfig := milvus2.IndexerConfig{
		Client:     newClient,                 //Milvus 客户端实例
		Collection: "personal_knowledge_base", //目标集合名称（注意：和 retriever 示例用的 "test" 不同，跑通前后端要注意一致）
		Embedding:  embedder,                  //Embedding 向量化实例
		Vector: &milvus2.VectorConfig{
			Dimension: 1024, // 向量维度，必须和 embedding 模型（bge-m3=1024维）一致，否则入库/检索会报维度不匹配
			MetricType: milvus2.COSINE, //距离度量方式：余弦相似度，最常用
			IndexBuilder: milvus2.NewHNSWIndexBuilder().WithM(16).WithEfConstruction(200), //HNSW 高性能向量索引
		},
	}

	indexer, err := milvus2.NewIndexer(ctx, &indexerConfig)
	if err != nil {
		panic(err)
	}
	// 1. 切分文档
	splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{
			"#":   "h1",
			"##":  "h2",
			"###": "h3",
		},
		TrimHeaders: false, // 保留标题，保证语义完整
	})
	if err != nil {
		panic(err)
	}

	content, err := os.ReadFile("个人成长与求职知识库.md")
	if err != nil {
		panic(err)
	}

	datas := []*schema.Document{
		{
			ID:      "doc_personal_knowledge",
			Content: string(content),
		},
	}

	transform, err := splitter.Transform(ctx, datas)
	if err != nil {
		panic(err)
	}

	// 2. 加工分块：过滤无效块 + 拼接标题路径 + 补充元数据
	var validChunks []*schema.Document
	for i, doc := range transform {
		content := strings.TrimSpace(doc.Content)
		// 过滤过短的无效块
		if len(content) < 20 {
			continue
		}

		// 从 metadata 提取各级标题，拼接完整路径
		h1, _ := doc.MetaData["h1"].(string)
		h2, _ := doc.MetaData["h2"].(string)
		h3, _ := doc.MetaData["h3"].(string)

		titlePath := h1
		if h2 != "" {
			titlePath += " > " + h2
		}
		if h3 != "" {
			titlePath += " > " + h3
		}

		// 补充元数据，标题路径前置到正文，一起向量化
		doc.ID = fmt.Sprintf("chunk_%03d", i)
		doc.MetaData["title_path"] = titlePath
		doc.MetaData["chunk_index"] = i
		doc.Content = titlePath + "\n" + content

		validChunks = append(validChunks, doc)
	}

	fmt.Printf("切分完成，有效分块 %d 个，开始写入 Milvus\n", len(validChunks))

	// 3. 一次性入库
	ids, err := indexer.Store(ctx, validChunks)
	if err != nil {
		panic("存入 Milvus 失败：" + err.Error())
	}

	fmt.Println("全部完成，已存入向量 ID 列表:", ids)
}
