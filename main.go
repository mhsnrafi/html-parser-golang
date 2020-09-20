package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

type HtmlResponse struct {
	ID                     string         `json:"id"`
	URL                    string         `json:"url"`
	PageTitle              string         `json:"htmltitle"`
	HtmlVersion            string         `json:"htmlversion"`
	HeadingCount           map[string]int `json:"headingcount"`
	ExternalLinksCount     int            `json:"externallink"`
	InternalLinksCount     int            `json:"internalink"`
	InaccessibleLinksCount int            `json:"inaccessible"`
	IsLogin                bool           `json:"islogin"`
}

var htmlresponse []HtmlResponse

func getParserResponses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(htmlresponse)
}

func getParserResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	for _, item := range htmlresponse {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&HtmlResponse{})
}

func updateParserResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, item := range htmlresponse {
		if item.ID == params["id"] {
			htmlresponse = append(htmlresponse[:index], htmlresponse[index+1:]...)
			var resp HtmlResponse
			_ = json.NewDecoder(r.Body).Decode(&resp)
			resp.ID = params["id"] //Mock ID - not safe
			htmlresponse = append(htmlresponse, resp)
			return
		}
		json.NewEncoder(w).Encode(htmlresponse)
	}
}

func deleteParserResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, item := range htmlresponse {
		if item.ID == params["id"] {
			htmlresponse = append(htmlresponse[:index], htmlresponse[index+1:]...)
			break
		}
		json.NewEncoder(w).Encode(htmlresponse)
	}
}

func fetchParserResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var htmlresp HtmlResponse

	_ = json.NewDecoder(r.Body).Decode(&htmlresp)
	htmlresp = htmlParser(htmlresp.URL)
	htmlresp.ID = strconv.Itoa(rand.Intn(10))
	htmlresponse = append(htmlresponse, htmlresp)
	json.NewEncoder(w).Encode(htmlresponse)
}

func htmlParser(url string) HtmlResponse {
	// Request the HTML page.
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	html, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	pageContent := string(html)

	//Create new document
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	//Calling Requested Parameters
	pageTitle := getTitle(pageContent)
	islogin := checklogin(doc, pageContent)
	htmlVersion := getHtmlVersion(pageContent)
	linksCounts, externalLinks := getCountoflinks(doc)
	inAccessibleLinksCount := getCountOfInAccessibleLink(externalLinks)
	headingCountsByLevel := getHeadingCountsByLevel(doc)

	resp := HtmlResponse{
		URL:                    url,
		PageTitle:              pageTitle,
		HtmlVersion:            htmlVersion,
		ExternalLinksCount:     linksCounts["externallink"],
		InternalLinksCount:     linksCounts["internallink"],
		InaccessibleLinksCount: inAccessibleLinksCount,
		HeadingCount:           headingCountsByLevel,
		IsLogin:                islogin,
	}

	return resp
}

func getTitle(pageContent string) string {
	titleStartIndex := strings.Index(pageContent, "<title>")
	if titleStartIndex == -1 {
		fmt.Println("No title element found")
	}

	// <title> as part of the tag, so let's offset the index by the number of characers in <title>
	titleStartIndex += 7

	// Find the index of the closing tag
	titleEndIndex := strings.Index(pageContent, "</title>")
	if titleEndIndex == -1 {
		return "No closing tag for title found."
	}

	pageTitle := string([]byte(pageContent[titleStartIndex:titleEndIndex]))

	return pageTitle
}

func getHtmlVersion(pageContent string) string {
	doctTypeMap := make(map[string]string)

	//Html Versions Declaration
	doctTypeMap["html5"] = "<!doctype html>"
	doctTypeMap["HTML4.01-Strict"] = "<!doctype html public \"-//w3c//dtd html 4.01//en\">"
	doctTypeMap["HTML4.01-Transitional"] = "<!doctype html public \"-//w3c//dtd html 4.01 transitional//en\">"
	doctTypeMap["HTML4.01-Frameset"] = "<!doctype html public \"-//w3c//dtd html  4.01 frameset//en\">"

	for key, value := range doctTypeMap {
		if strings.Contains(strings.ToLower(pageContent), value) {
			return key
		}
	}
	return "No version found"
}

func getCountoflinks(doc *goquery.Document) (map[string]int, []string) {
	linkscountMap := make(map[string]int)
	var links []string
	var externalLinkCount, internalLinkCount int

	doc.Find("a[href]").Each(func(index int, item *goquery.Selection) {
		href, _ := item.Attr("href")
		if strings.Contains(href, "http://") || strings.Contains(href, "https://") {
			externalLinkCount += 1
			links = append(links, href)
		} else {
			internalLinkCount += 1
		}
	})

	linkscountMap["externallink"] = externalLinkCount
	linkscountMap["internallink"] = internalLinkCount

	return linkscountMap, links
}

func getCountOfInAccessibleLink(extlinks []string) int {
	count := 0
	for _, link := range extlinks {
		resp, err := http.Get(link)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode >= 399 {
			count += 1
		}
	}
	return count
}

func getHeadingCountsByLevel(doc *goquery.Document) map[string]int {
	count := 0
	var headingLevels = []string{"h1", "h2", "h3", "h4", "h5", "h6"}
	headingCount := make(map[string]int)

	for _, head := range headingLevels {
		doc.Find(head).Each(func(index int, item *goquery.Selection) {
			count += 1
			headingCount[head] = count
		})
		count = 0
	}

	return headingCount

}

func checklogin(doc *goquery.Document, pagecontent string) bool {
	flag := 0
	var str = []string{"sign in", "login"}

	if StringInSlice(strings.ToLower(getTitle(pagecontent)), str) {
		flag = 1
	} else {
		doc.Find("input").Each(func(i int, s *goquery.Selection) {
			name, ok := s.Attr("name")
			if ok {
				if strings.Contains(strings.ToLower(name), "password") {
					flag = 1
				}
			}
		})
	}

	if flag == 1 {
		return true
	}
	return false
}

//StringInSlice return true if string s is present in slice list, else false
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.Contains(a, b) {
			return true
		}
	}
	return false
}

func main() {
	//Init Router
	r := mux.NewRouter()

	//Mock Sample Data @todo - implement DB
	htmlresponse = append(htmlresponse, HtmlResponse{ID: "2", HtmlVersion: "html5", PageTitle: "Book One", ExternalLinksCount: 2, InternalLinksCount: 4, InaccessibleLinksCount: 5, IsLogin: false})

	//Route Handlers Endpoints
	r.HandleFunc("/api/response", getParserResponses).Methods("GET")
	r.HandleFunc("/api/response/{id}", getParserResponse).Methods("GET")
	r.HandleFunc("/api/response", fetchParserResponse).Methods("POST")
	r.HandleFunc("/api/response/{id}", updateParserResponse).Methods("PUT")
	r.HandleFunc("/api/response/{id}", deleteParserResponse).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8100", r))
}
