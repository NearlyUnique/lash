# Overview

The `lash` package is designed to write scripty type one of lash-up programs

## Simple Example

```golang
// read a file and for each line call a web API
void main() {
    lash.OnError(lash.Terminate)                // Terminate is a func that takes and error (this is default)

    for line := range lash.OpenRead("somefile").ReadLines() {
        for resp:= range lash.
            Curl("http://httpbin.org/post").
            Post().                             // default is GET
            Header("Any","Value").
            AuthBasic("username","password").   // formats the Authorization header for you
            BodyJSON(line).                     // Body([]byte]) or BodyJSON( anything that is json-able)
            ResponseChan() {                    // Response or ResponseChan
                fmt.Println(resp.Body.String())
        }
    }
}
```
