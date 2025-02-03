package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"time"
)

type hyper struct {
	Link        string
	Txt         string
	LastChecked time.Time
	Links       map[string]int
	Images      map[string]int
}

var stream []*item
var itemsMap map[string]*item = make(map[string]*item)

func itemView(id string) *item {
	return itemsMap[id]
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	data := partFormData(r, w)
	if data == nil {
		return
	}
	stream = append([]*item{data}, stream...)

	b, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
	}
	ajaxResponse(w, map[string]string{
		"success":    "true",
		"replyID":    data.ID,
		"itemString": string(b),
	})
	saveJSON()
}
func partFormData(r *http.Request, w http.ResponseWriter) *item {
	mr, err := r.MultipartReader()
	if err != nil {
		log.Println(err)
	}

	var data *item = &item{ID: genPostID(10)}

	data.Phrases = make(map[string]int)
	data.ImgsWithAlts = make(map[string][]string)
	data.ContentMap = make(map[string]int)
	data.Links = make(map[string]int)

	for {
		part, err_part := mr.NextPart()
		if err_part == io.EOF {
			break
		}
		if part.FormName() == "Link" {
			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(part)
			if err != nil {
				log.Println(err)
			}
			// p := buf.String()
			data.Link = buf.String()
			getData(data)
			getImageLinkAndAltText(data)
			getTitleAndContent(data)
			if data.Status != "content" {
				fmt.Println("Error:", err)
				ajaxResponse(w, map[string]string{
					"success": "false",
					"msg":     "bad link",
				})
				return nil
			}
		}
	}
	return data
}
func init() {
	readDB()
	// err := os.Mkdir("./public/temp", 0777)
	// if err != nil {
	// 	log.Println(err)
	// }
}

type database struct {
	Pages []*item `json:"Pages"`
}

var items []*item = []*item{}
var db *database = &database{Pages: items}

func readDB() {
	content, err := os.ReadFile("JSON_DB.json")
	if err != nil {
		log.Println(err)
	}

	if len(content) > 0 {
		err := json.Unmarshal(content, db)
		if err != nil {
			log.Println(err)
		}
		// db.Pages = db.Pages[:50]
		slices.Reverse(db.Pages)
		stream = append(stream, db.Pages...)

		for _, item := range stream {
			itemsMap[item.ID] = item
		}
	}
}

func saveJSON() {
	// f, err := os.OpenFile("JSON_DB.json", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	f, err := os.Create("JSON_DB.json")
	if err != nil {
		log.Println(err)
	}

	defer f.Close()

	var stream_ []*item = make([]*item, len(stream))
	copy(stream_, stream)
	slices.Reverse(stream_)
	b, err := json.Marshal(&database{Pages: stream_})
	if err != nil {
		log.Println(err)
	}

	if _, err = f.WriteString(string(b)); err != nil {
		log.Println(err)
	}

	readDB()
}

type item struct {
	FileElement  string    `json:"FileElement"`
	Link         string    `json:"Link"`
	ID           string    `json:"ID"`
	Submitted    time.Time `json:"submitted"`
	LastChecked  time.Time
	Status       string
	Title        string `json:"title"`
	Content      string `json:"content"`
	HTML         string
	Text         string `json:"text"`
	Image        string
	Links        map[string]int `json:"links"`
	Images       []string       `json:"images"`
	LastErr      string
	Phrases      map[string]int
	ContentMap   map[string]int
	ImgsWithAlts map[string][]string
}

// type item struct {
// 	Link        string    `json:"Link"`
// 	Say_IT      string    `json:"Say IT"`
// 	ID          string    `json:"ID"`
// 	TS          time.Time `json:"TS"`
// 	Status      string
// 	// StatusChan   chan any
// 	MediaType    string `json:"mediaType"`
// 	TempFileName string `json:"tempFileName"`
// 	Title        string `json:"title"`
// 	Text         string `json:"text"`
// 	HTML         string
// 	Image        string
// 	Content      []string `json:"content"`
// 	Links        []string `json:"links"`
// 	Images       []string `json:"images"`
// 	LastChecked  time.Time
// 	Submitted    time.Time `json:"submitted"`
// 	LastErr      string

// 	ImgsWithAlts map[string][]string
// }
