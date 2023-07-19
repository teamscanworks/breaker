# Breaker

`breaker` functions as a basic service for the `x/circuit` module, facilitating circuit breaker capabilities, exposing the ability to trip and reset circuits via a HTTP API.

# Features

* API driven
    * Trip and reset circuits
    * Fetch statistical information (disabled commands, etc..)
    * JWT authentication
* YAML based configuration 
* Service specific keyring

# Usage

For usage related documentation please consult the (docs folder)[./docs/README.md] 

## Running Tests

Due to the usage of `x/circuit`, a special test environment needs to be prepared in order to accurately run all tests. This can be done by running the following commands anytime you want to start from a fresh golden image:

```shell
$> make reset-simd
$> make start-simd
$> ./scripts/submit_prop.sh
$> ./scripts/submit_votes.sh # this sleeps for about 70 seconds to allow gov proposal to pass
```

After the above commands have completed you may now now run all unit tests:

```shell
$> make test
```