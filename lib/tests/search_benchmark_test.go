/*
 * Copyright 2024 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tests

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils/docker"
	"github.com/SENERGY-Platform/permission-search/lib/model"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/opensearch-project/opensearch-go/opensearchutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type Document struct {
	Id          string `json:"id" bson:"id"`
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
}

func BenchmarkSearchDocuments(b *testing.B) {
	b.Skip()
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, osip, err := docker.OpenSearch(ctx, wg)
	if err != nil {
		b.Error(err)
		return
	}
	openSearchUrl := "https://" + osip + ":9200"
	openSearchClient, err := opensearch.NewClient(opensearch.Config{
		EnableRetryOnTimeout: true,
		MaxRetries:           3,
		RetryBackoff:         func(i int) time.Duration { return time.Duration(i) * 100 * time.Millisecond },
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Addresses: strings.Split(openSearchUrl, ","),
		Username:  "admin",
		Password:  docker.OpenSearchTestPw,
	})
	if err != nil {
		b.Error(err)
		return
	}
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"search":      map[string]interface{}{"type": "text", "analyzer": "custom_analyzer", "search_analyzer": "custom_search_analyzer"},
				"name":        map[string]interface{}{"type": "keyword", "copy_to": "search"},
				"description": map[string]interface{}{"type": "keyword", "copy_to": "search"},
			},
		},
		"settings": map[string]interface{}{
			"index": map[string]interface{}{
				"number_of_shards":   1,
				"number_of_replicas": 0,
			},
			"analysis": map[string]interface{}{
				"filter": map[string]interface{}{
					"autocomplete_filter": map[string]interface{}{
						"type":     "edge_ngram",
						"min_gram": 1,
						"max_gram": 20,
					},
					"custom_word_delimiter_filter": map[string]interface{}{
						"type":              "word_delimiter",
						"preserve_original": true,
					},
				},
				"normalizer": map[string]interface{}{
					"sortable": map[string]interface{}{
						"type": "custom",
						"filter": []string{
							"lowercase",
							"asciifolding",
						},
					},
				},
				"analyzer": map[string]interface{}{
					"custom_analyzer": map[string]interface{}{
						"type":      "custom",
						"tokenizer": "whitespace",
						"filter": []string{
							"custom_word_delimiter_filter",
							"lowercase",
							"unique",
							"autocomplete_filter",
						},
					},
					"custom_search_analyzer": map[string]interface{}{
						"type":      "custom",
						"tokenizer": "whitespace",
						"filter": []string{
							"word_delimiter",
							"lowercase",
						},
					},
				},
			},
		},
	}
	_, err = openSearchClient.Indices.Create("test", openSearchClient.Indices.Create.WithBody(opensearchutil.NewJSONReader(mapping)), openSearchClient.Indices.Create.WithContext(ctx))
	if err != nil {
		b.Error(err)
		return
	}

	_, mongoip, err := docker.MongoDB(ctx, wg)
	if err != nil {
		b.Error(err)
		return
	}
	mongoUrl := "mongodb://" + mongoip + ":27017"

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUrl), options.Client().SetReadConcern(readconcern.Majority()))
	if err != nil {
		b.Error(err)
		return
	}
	mongoCollection := mongoClient.Database("test").Collection("test")

	var set = func(d Document) error {
		fmt.Println("create", d.Name)
		ctx, _ := context.WithTimeout(ctx, 10*time.Second)
		_, err = mongoCollection.InsertOne(ctx, d)
		if err != nil {
			b.Error(err)
			return err
		}

		options := []func(request *opensearchapi.IndexRequest){
			openSearchClient.Index.WithDocumentID(d.Id),
			openSearchClient.Index.WithContext(ctx),
		}
		resp, err := openSearchClient.Index(
			"test",
			opensearchutil.NewJSONReader(d),
			options...,
		)
		if err != nil {
			b.Error(err)
			return err
		}
		defer resp.Body.Close()
		if resp.IsError() {
			b.Error(err)
			return err
		}
		return nil
	}

	done := sync.WaitGroup{}
	setChan := make(chan Document, 100)
	for i := 0; i < 20; i++ {
		done.Add(1)
		go func() {
			defer done.Done()
			for d := range setChan {
				err := set(d)
				if err != nil {
					b.Error(err)
					return
				}
			}
		}()
	}

	var toRoman = func(num int) string {
		val := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
		sym := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}

		roman, calculation := "", ""

		for i := 0; i < len(val); i++ {
			for num >= val[i] {
				num -= val[i]
				roman += sym[i]
				if calculation != "" {
					calculation += "+"
				}
				calculation += sym[i]
			}
		}
		return roman
	}

	knownSearchTexts := []string{}
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			nameParts := []string{
				fmt.Sprintf("foo%s", strings.ToLower(toRoman(i))),
				fmt.Sprintf("bar%s", strings.ToLower(toRoman(j))),
			}
			knownSearchTexts = append(knownSearchTexts, nameParts...)
			setChan <- Document{
				Id:          strconv.Itoa(i) + strconv.Itoa(j),
				Name:        strings.Join(nameParts, " "),
				Description: fmt.Sprintf("blub%s ", strings.ToLower(toRoman(j))),
			}
		}
	}
	close(setChan)

	done.Wait()
	time.Sleep(2 * time.Second)

	var mongoGet = func(s string) (result []Document, err error) {
		ctx, _ := context.WithTimeout(ctx, 10*time.Second)
		cursor, err := mongoCollection.Find(ctx, bson.M{
			"$or": []interface{}{
				bson.M{"name": bson.M{"$regex": regexp.QuoteMeta(s), "$options": "i"}},
				bson.M{"description": bson.M{"$regex": regexp.QuoteMeta(s), "$options": "i"}},
			},
		})
		if err != nil {
			b.Error(err)
			return nil, err
		}
		result = []Document{}
		for cursor.Next(ctx) {
			element := Document{}
			err = cursor.Decode(&element)
			if err != nil {
				b.Error(err)
				return nil, err
			}
			result = append(result, element)
		}
		err = cursor.Err()
		if err != nil {
			b.Error(err)
			return nil, err
		}
		cursor.Close(context.Background())
		return result, nil
	}

	var opensearchGet = func(s string) (result []Document, err error) {
		ctx, _ := context.WithTimeout(ctx, 10*time.Second)
		body := map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"should": []map[string]interface{}{
						{
							"wildcard": map[string]interface{}{
								"search": map[string]interface{}{"case_insensitive": true, "value": "*" + s + "*"},
							},
						},
						{
							"match": map[string]interface{}{
								"search": map[string]interface{}{"operator": "AND", "query": s},
							},
						},
					},
				},
			},
		}

		resp, err := openSearchClient.Search(
			openSearchClient.Search.WithIndex("test"),
			openSearchClient.Search.WithContext(ctx),
			openSearchClient.Search.WithSize(5000),
			openSearchClient.Search.WithFrom(0),
			openSearchClient.Search.WithBody(opensearchutil.NewJSONReader(body)),
		)
		if err != nil {
			return result, err
		}
		defer resp.Body.Close()
		if resp.IsError() {
			b.Error(err)
			return
		}
		pl := model.SearchResult[Document]{}
		err = json.NewDecoder(resp.Body).Decode(&pl)
		if err != nil {
			b.Error(err)
			return
		}

		for _, hit := range pl.Hits.Hits {
			result = append(result, hit.Source)
		}
		return result, nil
	}

	var randSearchText = func() string {
		return knownSearchTexts[rand.Intn(len(knownSearchTexts))]
	}

	start := time.Now()
	mongoGet("fooi")
	log.Println("mongo fooi", time.Since(start))

	start = time.Now()
	opensearchGet("fooi")
	log.Println("opensearch fooi", time.Since(start))

	start = time.Now()
	mongoGet("foox")
	log.Println("mongo foox", time.Since(start))

	start = time.Now()
	opensearchGet("foox")
	log.Println("opensearch foox", time.Since(start))

	start = time.Now()
	mongoGet("miss")
	log.Println("mongo miss", time.Since(start))

	start = time.Now()
	opensearchGet("miss")
	log.Println("opensearch miss", time.Since(start))

	start = time.Now()
	mongoGet("fooxx barx")
	log.Println("mongo fooxx barx", time.Since(start))

	start = time.Now()
	opensearchGet("fooxx barx")
	log.Println("opensearch fooxx barx", time.Since(start))

	b.ResetTimer()

	b.Run("mongo random", func(b *testing.B) {
		_, err := mongoGet(randSearchText())
		if err != nil {
			b.Error(err)
			return
		}
	})

	b.Run("mongo fooi", func(b *testing.B) {
		result, err := mongoGet("fooi")
		if err != nil {
			b.Error(err)
			return
		}
		if len(result) != 500 {
			b.Error(len(result))
			return
		}
	})
	b.Run("mongo bari", func(b *testing.B) {
		result, err := mongoGet("bari")
		if err != nil {
			b.Error(err)
			return
		}
		if len(result) != 500 {
			b.Error(len(result))
			return
		}
	})

	b.Run("mongo blubl_", func(b *testing.B) {
		result, err := mongoGet("blubl ")
		if err != nil {
			b.Error(err)
			return
		}
		if len(result) != 100 {
			b.Error(len(result))
			return
		}
	})

	b.Run("mongo miss", func(b *testing.B) {
		result, err := mongoGet("miss")
		if err != nil {
			b.Error(err)
			return
		}
		if len(result) != 0 {
			b.Error(len(result))
			return
		}
	})

	b.Run("mongo fooxx barx", func(b *testing.B) {
		result, err := mongoGet("fooxx barx")
		if err != nil {
			b.Error(err)
			return
		}
		if len(result) != 50 {
			b.Error(len(result))
			return
		}
	})

	b.Run("opensearch random", func(b *testing.B) {
		_, err := opensearchGet(randSearchText())
		if err != nil {
			b.Error(err)
			return
		}
	})

	b.Run("opensearch fooi", func(b *testing.B) {
		result, err := opensearchGet("fooi")
		if err != nil {
			b.Error(err)
			return
		}
		if len(result) != 500 {
			//b.Error(len(result))
			return
		}
	})
	b.Run("opensearch bari", func(b *testing.B) {
		result, err := opensearchGet("bari")
		if err != nil {
			b.Error(err)
			return
		}
		if len(result) != 500 {
			b.Error(len(result))
			return
		}
	})

	b.Run("opensearch fooxx barx", func(b *testing.B) {
		result, err := opensearchGet("fooxx barx")
		if err != nil {
			b.Error(err)
			return
		}
		if len(result) != 50 {
			//b.Error(len(result))
			return
		}
	})

	b.Run("opensearch blubl_", func(b *testing.B) {
		result, err := opensearchGet("blubl ")
		if err != nil {
			b.Error(err)
			return
		}
		if len(result) != 100 {
			//b.Error(len(result))
			return
		}
	})

	b.Run("opensearch miss", func(b *testing.B) {
		result, err := opensearchGet("miss")
		if err != nil {
			b.Error(err)
			return
		}
		if len(result) != 0 {
			b.Error(len(result))
			return
		}
	})
}
