package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

type HtmlResponse struct {
	ID                     string `json:"id"`
	URL                    string `json:"url"`
	PageTitle              string `json:"htmltitle"`
	HtmlVersion            string `json:"htmlversion"`
	HeadingCount           int    `json:"headingcount"`
	ExternalLinksCount     int    `json:"externallink"`
	InternalLinksCount     int    `json:"internalink"`
	InaccessibleLinksCount int    `json:"inaccessible"`
	IsLogin                bool   `json:"islogin"`
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
	htmlresp = html_parser(htmlresp.URL)
	htmlresp.ID = strconv.Itoa(rand.Intn(10)) //Mock ID - not safe
	htmlresponse = append(htmlresponse, htmlresp)
	json.NewEncoder(w).Encode(htmlresponse)
}

func html_parser(url string) HtmlResponse {
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
	//Get Page Title
	pageTitle := getTitle(pageContent)

	//Calling Html Version
	htmlVersion := getHtmlVersion(pageContent)

	linksCounts, externalLinks := getCountoflinks(url)
	inAccessibleLinksCount := getCountOfInAccessibleLink(externalLinks)
	resp := HtmlResponse{
		URL:                    url,
		PageTitle:              pageTitle,
		HtmlVersion:            htmlVersion,
		ExternalLinksCount:     linksCounts["externallink"],
		InternalLinksCount:     linksCounts["internallink"],
		InaccessibleLinksCount: inAccessibleLinksCount,
	}

	return resp
}

func getTitle(pageContent string) string {
	// Find a title
	titleStartIndex := strings.Index(pageContent, "<title>")
	if titleStartIndex == -1 {
		fmt.Println("No title element found")
		os.Exit(0)
	}
	// The start index of the title is the index of the first
	// character, the < symbol. We don't want to include
	// <title> as part of the final value, so let's offset
	// the index by the number of characers in <title>
	titleStartIndex += 7

	// Find the index of the closing tag
	titleEndIndex := strings.Index(pageContent, "</title>")
	if titleEndIndex == -1 {
		fmt.Println("No closing tag for title found.")
		os.Exit(0)
	}

	// (Optional)
	// Copy the substring in to a separate variable so the
	// variables with the full document data can be garbage collected
	pageTitle := []byte(pageContent[titleStartIndex:titleEndIndex])

	// Print out the result
	return string(pageTitle)

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

func getCountoflinks(pageurl string) (map[string]int, []string) {
	linkscountMap := make(map[string]int)
	var links []string
	doc, err := goquery.NewDocument(pageurl)
	var externalLinkCount, internalLinkCount int
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("a[href]").Each(func(index int, item *goquery.Selection) {
		href, _ := item.Attr("href")
		//fmt.Printf("link: %s - anchor text: %s\n", href, item.Text())
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

func main() {
	//Init Router
	r := mux.NewRouter()

	//Mock Data @todo - implement DB
	htmlresponse =
		append(htmlresponse,
			HtmlResponse{ID: "2", HtmlVersion: "html5", PageTitle: "Book One", HeadingCount: 5},
		)

	//Route Handlers  /Exnpoints
	r.HandleFunc("/api/response", getParserResponses).Methods("GET")
	r.HandleFunc("/api/response/{id}", getParserResponse).Methods("GET")
	r.HandleFunc("/api/response", fetchParserResponse).Methods("POST")
	r.HandleFunc("/api/response/{id}", updateParserResponse).Methods("PUT")
	r.HandleFunc("/api/response/{id}", deleteParserResponse).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8100", r))
}
