package main

import (
	"fmt"
	"net/http"
	"strings"
)

func viewItem(w http.ResponseWriter, r *http.Request) {
	// readDB()
	id := strings.Split(r.RequestURI, "/")[2]
	fmt.Println(r.RequestURI, id, itemsMap[id], itemsMap)
	exeTmpl(w, r, &viewData{AppName: appConf.App.Name, Stream: []*item{itemsMap[id]}}, "main.html")
}
