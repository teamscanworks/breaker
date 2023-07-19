# CLI

## Build CLI

To build the CLI run `make build`, which outputs an executable `breaker-cli` in the current directory.

## Populate Configuration File

To populate the config file with a default configuration suitable for further customization run:

```shell
$> ./breaker-cli config new
```

## Initialize Keyring


### Creating New Mnemonic

When initializing the configuration file for the first time it is recommended that you create a new mnemonic which will be used to facilitate the actual signing of transactions.

To do so run the following command, which avoids logging the mnemonic phrase by using `fmt.Println` instead.


```shell
$> ./breaker-cli config new-key --create.mnemonic --key.name <SIGNING_KEY_NAME>
{"level":"info","ts":1688684424.1259928,"logger":"compass","caller":"compass@v0.0.0-20230706013203-3802c4524650/client.go:92","msg":"initialized client"}
{"level":"warn","ts":1688684424.1260705,"logger":"breaker.client","caller":"breakerclient/breakerclient.go:94","msg":"no keys found, you should create at least one"}
{"level":"info","ts":1688684424.126079,"caller":"cli/cli.go:216","msg":"creating mnemonic"}
Enter keyring passphrase (attempt 1/3):
Re-enter keyring passphrase:
mnemonic  icon concert service unusual wonder observe radar flock other lunch antique patch company snack gravity invest hurt seek card mercy point gadget legal violin
```

### Importing Pre-Existing Mnemonic

If you want to import a pre-existing mnemonic instead of creating one, simply omit the `--create.mnemonic` flag, and you will be prompted to enter in your mnemonic.

```shell
{"level":"info","ts":1689799590.4641514,"logger":"compass","caller":"compass@v0.0.1/client.go:129","msg":"initialized client"}
{"level":"info","ts":1689799590.4642184,"caller":"cli/cli.go:221","msg":"reading mnemonic from user input"}
please paste your mnemonic phrase
{"level":"info","ts":1689799591.4337661,"logger":"breaker.client","caller":"breakerclient/breakerclient.go:153","msg":"keyring migration ok"}
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
$> ./breaker-cli api start --key.name <SIGNING_KEY_NAME>
{"level":"info","ts":1688668051.7295012,"logger":"compass","caller":"compass@v0.0.0-20230706013203-3802c4524650/client.go:92","msg":"initialized client"}
Enter keyring passphrase (attempt 1/3):
{"level":"info","ts":1688668053.498679,"logger":"breaker.client","caller":"breakerclient/breakerclient.go:97","msg":"configured from address","from.address":"cosmos18q2gyed58368mmrkz3k30s6kyrx0p4wrykals7"}
```
