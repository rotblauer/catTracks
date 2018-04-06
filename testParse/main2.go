package main

import (
    "encoding/json"
    "log"
    "net/http"
    "time"

    "github.com/creack/httpreq"
)

// Req is the request query struct.
type Req struct {
    Fields    []string
    Limit     int
    Page      int
    Timestamp time.Time
}

func handler(w http.ResponseWriter, req *http.Request) {
    if err := req.ParseForm(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    data := &Req{}
    if err := (httpreq.ParsingMap{
        {Field: "limit", Fct: httpreq.ToInt, Dest: &data.Limit},
        {Field: "page", Fct: httpreq.ToInt, Dest: &data.Page},
        {Field: "fields", Fct: httpreq.ToCommaList, Dest: &data.Fields},
        {Field: "timestamp", Fct: httpreq.ToTSTime, Dest: &data.Timestamp},
    }.Parse(req.Form)); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    _ = json.NewEncoder(w).Encode(data)
}

func main() {
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
    // curl 'http://localhost:8080?timestamp=1437743020&limit=10&page=1&fields=a,b,c'
}
