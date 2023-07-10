# Breaker

`breaker` functions as a basic service for the `x/circuit` module, facilitating circuit breaker capabilities. Using a HTTP API payloads can be submitted to an endpoint that can be used to trip or reset circuits.

Given that `breaker` is limited by the functionality present in `x/circuit`, the ability to gate access to module request urls is limited to an allowed/denied list that applies to all addresses.

# Features

* API driven
    * Trip and reset circuits
    * Fetch statistical information (disabled commands, etc..)
    * JWT authentication
* YAML based configuration 
* Basic keyring management for cosmos-sdk

# Dependency Management

Until `compass` is publicly released, managing dependencies for `breaker` requires marking the compass repository as a private module, which can be done in on the following ways:

* running `go env GOPRIVATE=github.com/teamscanworks/compass`
* running `export GOPRIVATE=github.com/teamscanworks/compass`

# Usage

## Build CLI

To build the CLI run `make build`, which outputs an executable `breaker-cli` in the current directory.

## Populate Configuration File

To populate the config file with a default configuration suitable for further customization run:

```shell
$> ./breaker-cli config new
```

## Initialize Keyring

When initializing the configuration file for the first time it is recommended that you create a new mnemonic which will be used to facilitate the actual signing of transactions.

To do so run the following command, which avoids logging the mnemonic phrase by using `fmt.Println` instead.


```shell
$> ./breaker-cli config new-key --create.mnemonic
{"level":"info","ts":1688684424.1259928,"logger":"compass","caller":"compass@v0.0.0-20230706013203-3802c4524650/client.go:92","msg":"initialized client"}
{"level":"warn","ts":1688684424.1260705,"logger":"breaker.client","caller":"breakerclient/breakerclient.go:94","msg":"no keys found, you should create at least one"}
{"level":"info","ts":1688684424.126079,"caller":"cli/cli.go:216","msg":"creating mnemonic"}
Enter keyring passphrase (attempt 1/3):
Re-enter keyring passphrase:
mnemonic  icon concert service unusual wonder observe radar flock other lunch antique patch company snack gravity invest hurt seek card mercy point gadget legal violin
```

## Display Active Keypair

To display the active keypair that is used for signing transactions run the following command, making sure the address has appropriate permissions for using the `x/circuit` module

```shell
$> ./breaker-cli config list-active-keypair
{"level":"info","ts":1688684300.6084576,"logger":"compass","caller":"compass@v0.0.0-20230706013203-3802c4524650/client.go:92","msg":"initialized client"}
Enter keyring passphrase (attempt 1/3):
{"level":"info","ts":1688684302.6615007,"logger":"breaker.client","caller":"breakerclient/breakerclient.go:96","msg":"configured from address","from.address":"cosmos10d2kehl8ss0q5yn90hk4rw3fxk8nfug4saczwv"}
{"level":"info","ts":1688684302.6652818,"caller":"cli/cli.go:183","msg":"found active keypair","address":"cosmos10d2kehl8ss0q5yn90hk4rw3fxk8nfug4saczwv"}
```

## Running The API Server

After populating the configuration file you can start the API server as follows. You will be prompted to enter a password to decrypt the keyring that was previously configured.

```shell
$> ./breaker-cli api start                                                                                                                            11:27:31
{"level":"info","ts":1688668051.7295012,"logger":"compass","caller":"compass@v0.0.0-20230706013203-3802c4524650/client.go:92","msg":"initialized client"}
Enter keyring passphrase (attempt 1/3):
{"level":"info","ts":1688668053.498679,"logger":"breaker.client","caller":"breakerclient/breakerclient.go:97","msg":"configured from address","from.address":"cosmos18q2gyed58368mmrkz3k30s6kyrx0p4wrykals7"}
```

## Running Tests

To create a fresh test environment run `make reset-simd`. After this you can run `make start-simd` which will start a simd environment designed for basic testing of `breaker` and running unit tests.

```shell
$> make reset-simd
$> make start-simd
# open new terminal window
$> make test