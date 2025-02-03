package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var complete *database = &database{Pages: []*item{}}
var pages []*item = []*item{}
var checkedLast map[string]*hyper = make(map[string]*hyper)

// func cycleRead() {
// 	for {
// 		for i, l := range db.Pages {
// 			fmt.Print("LINK #: ", i+1, " of ", len(db.Pages), " ", l.Status, " ")
// 			switch l.Status {
// 			case "not started":
// 				if time.Since(l.LastChecked) >= 2*time.Hour {
// 					fmt.Print(" STARTING ", l.Link)
// 					getData(l)
// 				}
// 			case "downloaded":
// 				fmt.Print(" GROKING ", l.Link)
// 				getContent(l)
// 			case "content":
// 				fmt.Print(" MAPPING ", l.Link)
// 				mapout(l)
// 			case "complete":
// 				if len(l.Title) >= 3 && len(l.Content) < 1000 {
// 					complete.Pages = append(complete.Pages, l)
// 				}
// 				l.LastChecked = time.Now()
// 				l.Status = "not started"
// 			case "download failed":
// 				fmt.Println(l.LastErr)
// 				l.LastChecked = time.Now()
// 				l.Status = "not started"
// 			}
// 			fmt.Println(" ENDED ")
// 		}
// 		writeJSON(complete, "compl.json")
// 	}
// }

func writeJSON(d *database, fn string) {
	b, err := json.MarshalIndent(d, "", "    ")
	if err != nil {
		log.Println(err)
	}
	appendFile(string(b), fn)
}

func readdb() []*item {
	b_, err := os.ReadFile("news.json")
	if err != nil {
		log.Println(err)
		return nil
	}
	err = json.Unmarshal(b_, db)
	if err != nil {
		log.Println(err)
		return nil
	}
	log.Println(db.Pages)
	return db.Pages
}

func readCSV() {
	if readdb() == nil {
		os.Exit(0)
		b, err := os.ReadFile("news.csv")
		if err != nil {
			log.Println(err)
		}
		links := []string{}
		for _, l := range strings.Split(string(b), "\n") {
			links = append(links, strings.Split(l, ",")[0])
		}

		for _, l := range links {
			p := &item{}
			p.Status = "not started"
			p.Link = l
			pages = append(pages, p)
		}
		writeJSON(db, "news.json")
	}
}

func getRes(p *item) *http.Response {
	client := http.Client{
		Timeout: 4 * time.Second,
	}
	res, err := client.Get(p.Link)
	if err != nil {
		log.Println(err)
		p.LastErr = fmt.Sprint(err)
		return nil
	}
	return res
}
func getTitleAndContent(p *item) *item {
	for _, l := range strings.Split(p.HTML, "content=\"") {
		c := strings.Split(l, "\"")[0]
		if !strings.ContainsAny(c, "/><=") && strings.Count(c, " ") > 1 {
			if p.Content == "" {
				p.Content = c
			}
			p.ContentMap[c] = 1
		}
	}

	title_ := strings.Split(strings.Split(p.HTML, "</title>")[0], ">")
	p.Title = html.UnescapeString(title_[len(title_)-1])
	p.Status = "content"
	return p
}
func getData(p *item) *item {
	l := p.Link
	p.LastChecked = time.Now()
	res := getRes(p)
	if res == nil || res.Status != "200 OK" {
		p.Status = "download failed"
		return p
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		p.Status = "download failed"
		return p
	}
	if checkedLast[l] == nil {
		checkedLast[l] = &hyper{}
		checkedLast[l].Images = make(map[string]int)
		checkedLast[l].Links = make(map[string]int)
	}

	p.HTML = string(b)
	p.Submitted = time.Now()
	p.Status = "downloaded"

	// downloadIMG(p)
	return p
}

// var types []string = []string{".png", ".jpg", ".jpeg", ".webp", ".webm", ".mp4"}
var props []string = []string{"src=\"", "alt=\"", "srcset=\""}

func getImageLinkAndAltText(p *item) {
	imgs := strings.Split(strings.ReplaceAll(p.HTML, "<img", "\n<img"), "\n")
	if len(imgs) > 0 {
		for _, im := range imgs {
			var alt string
			var set []string
			for _, pr := range props {
				if strings.Contains(im, props[0]) && strings.Contains(im, props[1]) {
					s := strings.Split(strings.Split(im, ">")[0], pr)
					if len(s) > 1 {
						switch pr {
						case props[2]:
							set = append(set, strings.Split(strings.Split(s[1], "\"")[0], ",")...)
						case "src=\"":
							if p.Image == "" {
								p.Image = strings.Split(strings.Split(s[1], "\"")[0], "?")[0]
							}
							set = append(set, strings.Split(strings.Split(s[1], "\"")[0], "?")[0])
						case "alt=\"":
							alt = strings.Split(strings.Split(s[1], "\"")[0], "?")[0]
						}
					}
				}
			}

			for _, sst := range set {
				if p.ImgsWithAlts[alt] == nil {
					p.ImgsWithAlts[alt] = []string{}
				}
				p.ImgsWithAlts[alt] = append(p.ImgsWithAlts[alt], strings.Split(sst, "?")[0])
			}
		}
	}
}

type ialt struct {
	Image string
	Alt   string
}

func getImgAndAlt(s string) {
	s = strings.Join(strings.Fields(s), " ")
	s = strings.ReplaceAll(s, "<img", "\n<img")
	// var ia []*ialt = []*ialt{}
	for _, s_ := range strings.Split(s, "\n<img") {
		log.Println(strings.Split(s_, ">"))
		// ia[i].Alt := strings.Split(s, ">")[0]
	}
}

// func mapout(p *item) {
// 	checkedLast = make(map[string]*hyper)
// 	checkedLast[p.Link] = &hyper{}
// 	checkedLast[p.Link].Links = make(map[string]int)
// 	checkedLast[p.Link].Images = make(map[string]int)
// 	for _, s := range linePics(p.HTML, p.Link) {
// 		u_ := strings.Split(strings.Split(strings.Split(s, "\"")[0], " ")[0], "'")[0]
// 		u_ = strings.ReplaceAll(u_, ":", "_")
// 		if len(u_) > 10 && len(u_) < 600 && !doesntContain(u_) {
// 			u_ = decu(u_)
// 			u_ = strings.ReplaceAll(u_, " ", "")
// 			u, err := url.Parse(u_)
// 			if err != nil {
// 				log.Println(err, u)
// 				return
// 			}
// 			isimage := false
// 			s = strings.ReplaceAll(u.Hostname()+u.EscapedPath(), "_", ":")
// 			for _, t := range types {
// 				if strings.Contains(s, t) && !strings.Contains(s, " ") {
// 					isimage = true
// 					checkedLast[p.Link].Images[s] = checkedLast[p.Link].Images[s] + 1
// 				}
// 			}
// 			if !isimage {
// 				checkedLast[p.Link].Links[s] = checkedLast[p.Link].Links[s] + 1
// 			}
// 		}
// 	}
// 	for _, l := range checkedLast {
// 		for img := range l.Images {
// 			p.Images = append(p.Images, img)
// 		}
// 		for link := range l.Links {
// 			p.Links = append(p.Links, link)
// 		}

//		}
//		p.Status = "complete"
//	}
func doesntContain(u_ string) bool {
	return !strings.Contains(u_, "google") &&
		!strings.ContainsAny(u_, ";") &&
		!strings.ContainsAny(u_, "min.js") &&
		!strings.ContainsAny(u_, "wix-thunder") &&
		!strings.ContainsAny(u_, "wp-admin") &&
		!strings.ContainsAny(u_, "wp-content") &&
		!strings.ContainsAny(u_, "facebook") &&
		!strings.ContainsAny(u_, ".css") &&
		!strings.ContainsAny(u_, ".js") &&
		!strings.ContainsAny(u_, "BuildQuery")
}
func decu(encodedURL string) (decodedURL string) {
	decodedURL, err := url.QueryUnescape(encodedURL)
	if err != nil {
		return ""
	}
	return
}
func downloadIMG(it *item) {
	for _, l := range it.Images {
		if strings.Contains(l, ".jpg") {
			r, err := http.Get(l)
			if err != nil {
				log.Println(err)
				return
			}
			defer r.Body.Close()
			spl := strings.Split(l, "/")
			path := "public/media/" + spl[len(spl)-1]
			f, err := os.Create(path)
			if err != nil {
				log.Println(err)
				return
			}
			defer f.Close()
			it.Image = path
			_, err = f.ReadFrom(r.Body)
			if err != nil {
				log.Println(err)
			}
			break
		}
	}
}

var types []string = []string{".png", ".jpg", ".jpeg", ".webp", ".webm", ".mp4"}

func linePics(s, l string) []string {
	s = strings.Join(strings.Fields(s), " ")
	// s = strings.ReplaceAll(s, "\\\"", "\"")

	for _, t := range types {
		s = strings.ReplaceAll(s, t, t+"\n")
	}
	s = strings.ReplaceAll(s, "http", "\nhttp")
	s = strings.ReplaceAll(s, "url(", "\n"+l)
	s = strings.ReplaceAll(s, "href=\"/", "\n"+l+"/")

	// fmt.Println(strings.Split(s, "\n"), l)
	return strings.Split(s, "\n")
}

var contentmap map[string]int = make(map[string]int)

// func getContent(p *item) *item {
// 	hypt := p.HTML
// 	for _, l := range strings.Split(hypt, "content=\"") {
// 		c := strings.Split(l, "\"")[0]
// 		if !strings.ContainsAny(c, "/><=") && strings.Count(c, " ") > 1 {
// 			contentmap[c] = 1
// 		}
// 	}
// 	title_ := strings.Split(strings.Split(hypt, "</title>")[0], ">")
// 	p.Title = title_[len(title_)-1]
// 	txt_ := ""
// 	for _, txt := range detectText(hypt) {
// 		txt_ = txt_ + " " + txt
// 	}
// 	p.Text = strings.ReplaceAll(strings.Join(strings.Fields(txt_), " "), ".", ". ")
// 	for con := range contentmap {
// 		p.Content = append(p.Content, con)
// 	}
// 	p.Status = "content"
// 	return p
// }

func appendFile(l, fn string) {
	f, err := os.OpenFile(fn, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Println(err)
	}

	defer f.Close()

	if _, err = f.WriteString(l); err != nil {
		log.Println(err)
	}

}
func detectText(hypt string) []string {
	apen := []string{}
	hypt = strings.ReplaceAll(hypt, "&nbsp;", "")
	hypt = strings.ReplaceAll(strings.Join(strings.Fields(hypt), " "), "<", "\n<")
	for _, l := range strings.Split(hypt, ">") {
		if len(l) >= 2 {
			if string(l[0]) != "<" && string(l[1]) != "<" {
				c :=
					strings.Count(l, "{") + strings.Count(l, "}") +
						strings.Count(l, ";") + strings.Count(l, "/") +
						strings.Count(l, ".") + strings.Count(l, "\\")
				if c < len(l)/20 {
					apen = append(apen, strings.Join(strings.Fields(strings.Split(l, "<")[0]), " "))
				}
			}
		}
	}
	return apen
}
