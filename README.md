# Ev

> E(nvironment) v(ariable)

Tiny Go library to create typed environment variables.

```bash
go get github.com/metafates/ev
```

## Example

```go
const MyEnvVar ev.Var[int] = "MY_ENV_VAR"

const Verbose ev.Var[bool] = "VERBOSE"

func main() {
    // assume we have the following variables set
    os.Setenv("MY_ENV_VAR", "42")
    os.Setenv("VERBOSE", "true")

    if Verbose.Get() {
        n := MyEnvVar.Get()

        fmt.Println(n + n) // prints 84
    }
}
```

Supported types:

``` go
type Constraint interface {
    int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64
}
```

Values are parsed using `fmt.Sscanf`.
