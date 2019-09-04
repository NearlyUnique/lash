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

## Start with new scope

```go
scope := lash.NewScope()
```

A scope is for errors and what happen with those errors. Use `scope.OnError` to control the behaviour. There are build in functions or supply your own.

- `lash.Terminate`
- `lash.Ignore`
- `lash.Warn`

### String interpolation

All methods, where appropriate, will support environment variable interpolation. The `scope.EnvStr` is also available directly

```go
os.Setenv("my_variable", "banana")

s = scope.EnvStr("$user is my name other args can be passed in $1, $2, $3", 99, true, "any")

scope.OpenFile("my_variable.txt").AppendLine("$user as a line of text")

scope.Prinfln("hello $user you're in $pwd")
```

### Require conditions for command

If you need specific environment variables or arguments you can set these. If the requirements are not met you OnError func will be executed

```go
scope.Env().
    Require("var_name", "some description", "displayed if missing").
    Default("other_var", "default value")
scope.Args().
    Require(1, "some description for first argument, displayed if missing")
```

### Files

```go
f := scope.OpenFile("some/path")

// Reading
f.String() // read the content into a string
ch := f.ReadLines() // read the lines into an unbuffered channel
for line := ch {
    fmt.PrintLn(line)
}
// Writing
f.AppendLine("some text")
f.Truncate()

// get an appender to append lines from other go routines
appender := f.Appender()
go func(){
    appender.AppendLine("some text")
}()

// Close
f.Close()
```