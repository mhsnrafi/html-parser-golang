package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty"
	"github.com/stretchr/testify/assert"
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

func Test_StatusCodeShouldEqual200(t *testing.T) {

	client := resty.New()
	resp, _ := client.R().Get("http://localhost:8100/api/response/7")

	if resp.StatusCode() != 200 {
		t.Errorf("Unexpected status code, expected %d, got %d instead", 200, resp.StatusCode())
	}
}

func Test_ContentTypeShouldEqualApplicationJson(t *testing.T) {

	client := resty.New()
	resp, _ := client.R().Get("http://localhost:8100/api/response/7")

	assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
}

func Test_GetResponseShouldEqualToMockResponse(t *testing.T) {
	Client := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}
	//Here we call our api and get the json response
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8100/api/response/1", nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "parser-api")

	res, getErr := Client.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	response := HtmlResponse{}

	jsonErr := json.Unmarshal(body, &response)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	assert.Equal(t, "Juliana Bicycles | The Original Women's Mountain Bike", response.PageTitle)
	assert.Equal(t, "html5", response.HtmlVersion)
	assert.Equal(t, 59, response.ExternalLinksCount)
	assert.Equal(t, 170, response.InternalLinksCount)
	assert.Equal(t, 0, response.InaccessibleLinksCount)
	assert.Equal(t, 9, response.HeadingCount["h1"])
	assert.Equal(t, 41, response.HeadingCount["h2"])
	assert.Equal(t, 4, response.HeadingCount["h3"])
	assert.Equal(t, 4, response.HeadingCount["h4"])
	assert.Equal(t, false, response.IsLogin)
}