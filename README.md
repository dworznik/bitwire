# bitwire

[Bitwire](https://www.bitwire.co) command line client written in Go.

[Bitwire API documentation](https://developers.bitwire.co)

## Installation


### Installation from source


To install bitwire, run:

```
go get github.com/dworznik/bitwire
cd $GOPATH/github.com/dworznik/bitwire
make
```


### OS X binaries via [Homebrew](http://brew.sh)


Add the tap with `bitwire` formula before installing.

```
brew tap dworznik/tap
brew install bitwire
```


## CLI client usage

Bitwire client produces two kinds of output: text and JSON. The default is text. To enable JSON output, add `-j` switch.

Add `-s` switch, if want to use bitwire sandbox API.


For usage instruction, run:

```
bitwire
```


To set up API access credentials, run:
```
bitwire config
```

Listing transfers:

```
bitwire transfers
```

Listing recipients:
```
bitwire recipients
```

Displaying current exchange rates:
```
bitwire rates
```

Getting a list of available banks:
```
bitwire banks
```

Displaying account's limits
```
bitwire limits
```


### Working with JSON output in the shell


Add `-j` switch, to get the JSON API response object.

```
bitwire -j  transfers
```


Use [jq command-line JSON processor](https://stedolan.github.io/jq/) to manipulate the output JSON.


For example, to get a total amount of your completed transfers, run:

```
bitwire -j  transfers | jq  'map(select(.status == "PAID_COMPLETED") .amount|tonumber ) | add'
```


## Client library usage


### Initialization

Import the package:

```
import "github.com/dworznik/bitwire"
```

It is possible to create a client object with no API credentials.

```
mode := bitwire.PRODUCTION
client, err := bitwire.New(mode)
if err != nil {
  panic(err)
}
```

Until the client has been authenticated, only API methods that are not account specific re available.


### Authentication


To authenticate using username and password, use `LoginCredentials`.


```
creds := bitwire.Credentials{clientId, clientSecret, "password"}
login := bitwire.LoginCredentials{creds, username, password}
token, err := client.Authenticate(login)
```

To authenticate using an existing token, create a new client with `NewWithToken()` and the `Token` struct returned by `Authenticate()`.

```
mode := bitwire.PRODUCTION
client, err := bitwire.NewWithToken(token)
if err != nil {
  panic(err)
}
```

To be able to refresh the token, you need to use `NewWithConfig()` and provide `client_id` and `client_secret` along with the token.

```
mode := bitwire.PRODUCTION
client, err := bitwire.NewWithConfig(conf)
if err != nil {
  panic(err)
}
```


## TODO
  - Clean up the code
  - More docs
  - Getting account name
  - Getting recipient by id
  - Creating a recipient
  - Creating a transfer
  - Localization
  - Displaying a transaction QR code in the terminal
