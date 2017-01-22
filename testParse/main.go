package main
 
import (
        "fmt"
        "log"
        "net/http"
        "net/url"
        "strconv"
)
 
// conversion helpers
func parseBool(s string, dest interface{}) error {
        d, ok := dest.(*bool)
        if !ok {
                return fmt.Errorf("wrong type for parseBool: %T", dest)
        }
        // assume error = false
        *d, _ = strconv.ParseBool(s)
        return nil
}
 
func parseInt(s string, dest interface{}) error {
        d, ok := dest.(*int)
        if !ok {
                return fmt.Errorf("wrong type for parseInt: %T", dest)
        }
        n, err := strconv.Atoi(s)
        if err != nil {
                return err
        }
        *d = n
        return nil
}
 
// parsingMap
type parsingMap []parsingMapElem
 
type parsingMapElem struct {
        Field string
        Fct   func(string, interface{}) error
        Dest  interface{}
}
 
func (p parsingMap) parse(form url.Values) error {
        for _, elem := range p {
                if err := elem.Fct(elem.Field, elem.Dest); err != nil {
                        return err
                }
        }
        return nil
}
 
// http server
type query struct {
        Limit  int
        DryRun bool
}
 
func handler(w http.ResponseWriter, req *http.Request) {
        if err := req.ParseForm(); err != nil {
                log.Printf("Error parsing form: %s", err)
                return
        }
        q := &query{}
        if err := (parsingMap{
                {"limit", parseInt, &q.Limit},
                {"dryrun", parseBool, &q.DryRun},
        }).parse(req.Form); err != nil {
                log.Printf("Error parsing query string: %s", err)
                return
        }
 
        fmt.Fprintf(w, "hello world. Limit: %d, Dryrun: %t\n", q.Limit, q.DryRun)
}
 
func main() {
        log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(handler)))
}
