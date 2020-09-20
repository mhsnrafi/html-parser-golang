### GO Lang Html Parser
RESTful API to create, read, update and delete html response. No database implementation yet

### Quick Start
* Install `mux` mux router
* Install `goquery` from web
```
$ go get -u github.com/gorilla/mux
$ go get github.com/PuerkitoBio/goquery   
```

* Start the dev server: `go run main.go`
* Verify that the api server is up at: `http://localhost:8100`


### Editor
Please get **Golang**, anything else for *free* is a compromise that is just not worth.
Get started with trial version.



### Endpoints

### Get All Responses
GET `http://localhost:8100/api/response`:

### Get Single Response
GET `http://localhost:8100/api/response{id}`:

### Delete Response
DELETE `http://localhost:8100/api/response/{id}`:


### Create Response
POST `http://localhost:8100/api/response`:
```
# body parameter:
{
    "url": "https://www.julianabicycles.com/en-US""
}

Request sample:
{
    "id": "1",
    "url": "https://www.julianabicycles.com/en-US",
    "htmltitle": "Juliana Bicycles | The Original Women's Mountain Bike",
    "htmlversion": "html5",
    "headingcount": {
        "h1": 9,
        "h2": 41,
        "h3": 4,
        "h4": 4
    },
    "externallink": 59,
    "internalink": 170,
    "inaccessible": 0,
    "islogin": false
}
```

### Update Response
PUT `http://localhost:8100/api/response/{id}`:

```
# body parameter:
{
    "id": "3"
}


Request sample:
{
    "id": "3",
    "url": "https://www.julianabicycles.com/en-US",
    "htmltitle": "Juliana Bicycles | The Original Women's Mountain Bike",
    "htmlversion": "html5",
    "headingcount": {
        "h1": 9,
        "h2": 41,
        "h3": 4,
        "h4": 4
    },
    "externallink": 59,
    "internalink": 170,
    "inaccessible": 0,
    "islogin": false
}
```


