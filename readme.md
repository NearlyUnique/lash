# Overview

The `lash` package is designed to write scripty type one of lash-up programs

## Simple Example

```golang
// read a file and for each line call a web API
void main() {
    scope:=NewScope()
    scope.OnError(lash.Terminate)                // Terminate is a func that takes and error (this is default)

    for line := range scope.OpenFile("somefile").ReadLines() {
        response := scope.
            Curl("https://httpbin.org/post").
            Post(line).                             // default is GET
            Header("Any","Value").
            AuthBasic("username","password").   // formats the Authorization header for you
            Response()

        fmt.Println(response.JSONBody())
    }
}
```
