package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// User represents a simple document we index into Elasticsearch.
type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 1. Create Elasticsearch client
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}

	// Quick ping
	res, err := es.Info()
	if err != nil {
		log.Fatalf("es.Info: %v", err)
	}
	defer res.Body.Close()
	log.Println("connected to Elasticsearch")

	// 2. Index a user document
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := User{Name: "Cristi", Age: 35}

	docBytes, err := json.Marshal(user)
	if err != nil {
		log.Fatalf("marshal user: %v", err)
	}

	indexReq := esapi.IndexRequest{
		Index:      "users",
		DocumentID: "1",
		Body:       bytes.NewReader(docBytes),
		Refresh:    "true", // make it searchable immediately
	}

	indexRes, err := indexReq.Do(ctx, es)
	if err != nil {
		log.Fatalf("index request: %v", err)
	}
	defer indexRes.Body.Close()

	if indexRes.IsError() {
		body, _ := io.ReadAll(indexRes.Body)
		log.Fatalf("index error: %s", string(body))
	}

	log.Println("indexed user with id=1")

	// 3. Get the same document back
	getReq := esapi.GetRequest{
		Index:      "users",
		DocumentID: "1",
	}

	getRes, err := getReq.Do(ctx, es)
	if err != nil {
		log.Fatalf("get request: %v", err)
	}
	defer getRes.Body.Close()

	if getRes.IsError() {
		body, _ := io.ReadAll(getRes.Body)
		log.Fatalf("get error: %s", string(body))
	}

	var getBody struct {
		Source User `json:"_source"`
	}

	if err := json.NewDecoder(getRes.Body).Decode(&getBody); err != nil {
		log.Fatalf("decode get body: %v", err)
	}

	fmt.Printf("User from Elasticsearch: name=%s age=%d\n",
		getBody.Source.Name, getBody.Source.Age)
}
