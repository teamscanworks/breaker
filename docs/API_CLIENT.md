# API Client

The `api` package contains a HTTP client for the API that allows for full usage of the breaker api server.

# Usage

## Get A Valid JWT

The breaker API server protects access to circuit breaking routes via usage of a JWT. As a reminder the default settings of the JWT are using the `HS256` signing algorith with the password protecting the JWT derived from the yaml configuration file key `api.password`.

The `breaker-cli` provides a method that can be used to issue a JWT:

```shell
$> ./breaker-cli api issue-jwt
{"level":"info","ts":1689799819.8215516,"caller":"cli/cli.go:62","msg":"issued token","jwt.token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODk4ODYyMjAsImlhdCI6MTY4OTc5OTgxOX0.1zuiKz8Y6AXvXwht3Kv2J3O8hsWPxgoSzZ1dlKOfj5Y"}
```

As per the above output, the JWT is `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODk4ODYyMjAsImlhdCI6MTY4OTc5OTgxOX0.1zuiKz8Y6AXvXwht3Kv2J3O8hsWPxgoSzZ1dlKOfj5Y`.

## Construct API Client

```go
package main
import (
    "ctx"
    "github.com/teamscanworks/breaker/api"

)

func main() {

    apiClient := api.NewAPIClient("http://127.0.0.1:42690", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODk4ODYyMjAsImlhdCI6MTY4OTc5OTgxOX0.1zuiKz8Y6AXvXwht3Kv2J3O8hsWPxgoSzZ1dlKOfj5Y")

    // fetch a list of disabled commands (doesn't require a valid jwt)
    cmds, err := apiClient.DisabledCommands()
    //...

    // fetch a list of accounts which have some permission to the x/circuit module (doesnt require a valid jwt)
    accts, err := apiClient.Accounts()
    //...

    // trip a circuit, preventing access to the given urls (requires valid jwt)
    // the provided message is logged locally, to assist with debugging
    resp, err := apiClient.TripCircuit([]string{"/some/cosmos/url"}, "a message to log")
    // ..

    // reset a circuit, allowing access to the given urls (required valid jwt)
    // the provided message is logged locally, to assist with debugging
    resp, err := apiClient.ResetCircuit([]string{"/some/cosmos/url"}, "a message to log")
}

```