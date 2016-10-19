package bitwire

import (
  "errors"
  "github.com/dghubble/sling"
  "time"
)

const baseURL = "https://www.bitwire.co/api/v1/"
const sandboxBaseURL = "https://sandbox.bitwire.co/api/v1/"

type Res struct {
  Code int `json:"code"`
}

type ErrorRes struct {
  Res
  Error
}

type Error struct {
  Message   string `json:"message"`
  ErrorType string `json:"errorType"`
}

type AllRatesRes struct {
  Res
  Rates AllRates `json:"rates"`
}

type Rates map[string]string

type BtcRatesRes struct {
  Res
  Rates Rates `json:"rates"`
}

type FxRatesRes struct {
  Res
  Rates Rates `json:"rates"`
}

type AllRates struct {
  BTC Rates `json:"btc"`
  FX  Rates `json:"fx"`
}

type BanksRes struct {
  Res
  Banks []Bank `json:"banks"`
}

type Bank struct {
  Id          int    `json:"id"`
  Number      string `json:"number"`
  DisplayName string `json:"display_name"`
  Name        string `json:"name"`
  NameKo      string `json:"name_ko"`
}

type RecipientsRes struct {
  Res
  Recipients []Recipient `json:"recipients"`
}

type TransferRes struct {
  Res
  Transfer Transfer
}

type TransfersRes struct {
  Res
  Transfers []Transfer
}

type Transfer struct {
  Id        string            `json:"id"`
  Sender    Sender            `json:"sender"`
  Type      string            `json:"type"`
  Memo      string            `json:"memo"`
  Amount    string            `json:"amount"`
  Currency  string            `json:"currency"`
  Status    string            `json:"status"`
  Date      string            `json:"date"`
  BTC       BTC               `json:"btc"`
  Recipient TransferRecipient `json:"recipient"`
}

type CreateTransfer struct {
  Amount      string `json:"amount"`
  Currency    string `json:"currency"`
  RecipientId int    `json:"recipient_id"`
  Memo        string `json:"memo"`
  Type        string `json:"type"`
}

type Sender struct {
  Amount   string `json:"amount"`
  Currency string `json:"currency"`
}

type Recipient struct {
  Id    int           `json:"id"`
  Name  string        `json:"name"`
  Email string        `json:"email"`
  Bank  RecipientBank `json:"bank"`
}

type TransferRecipient struct {
  Recipient
  Currency string `json:"currency"`
  Amount   string `json:"amount"`
}

type BTC struct {
  Address    string `json:"address"`
  Link       string `json:"link"`
  Expiration int    `json:"expiration"`
}

type RecipientBank struct {
  Bank
  AccountNumber string `json:"account_number"`
  AccountName   string `json:"account_name"`
}

type LimitsRes struct {
  Res
  Limits Limits `json:"limits"`
}

type Limits struct {
  Transfers TransferLimits `json:"transfers"`
  KRW       struct {
    Min    string    `json:"min"`
    Daily  KrwLimits `json:"daily"`
    Weekly KrwLimits `json:"weekly"`
  } `json:"krw"`
  BTC struct {
    Min string `json:"min"`
  }
}

type KrwLimits struct {
  Used  string `json:"used"`
  Left  string `json:"left"`
  Limit string `json"limit"`
}

type TransferLimits struct {
  Pending struct {
    Total struct {
      Used  int `json:"used"`
      Limit int `json:"limit"`
    } `json:"total"`
  } `json:"pending"`
  Completed struct {
    Daily struct {
      Used  int `json:"used"`
      Limit int `json:"limit"`
    } `json:"daily"`
  } `json:"completed"`
}

type TokenRes struct {
  Res
  Token
}

type Token struct {
  TokenType    string `json:"token_type"`
  AccessToken  string `json:"access_token"`
  RefreshToken string `json:"refresh_token"`
  ExpiresIn    int    `json:"expires_in"`
  ValidUntil   int64  `json:"valid_until"`
}

type Mode string

const (
  PRODUCTION Mode = "production"
  SANDBOX    Mode = "sandbox"
)

type Credentials struct {
  ClientId     string `json:"client_id" url:"client_id"`
  ClientSecret string `json:"client_secret" url:"client_secret"`
  GrantType    string `json:"grant_type" url:"grant_type"`
}

type LoginCredentials struct {
  Credentials
  Username string `url:"username"`
  Password string `url:"password"`
}

type TokenCredentials struct {
  Credentials
  RefreshToken string `json:"refresh_token" url:"refresh_token"`
}

type Config struct {
  Credentials
  Token Token `json:"token"`
}

type Client struct {
  Mode        Mode
  token       Token
  credentials Credentials
}

type Method string

const (
  GET       Method = "GET"
  POST      Method = "POST"
  JSON_POST Method = "JSON_POST"
  DELETE    Method = "DELETE"
)

func New(mode Mode) (*Client, error) {
  return NewWithToken(mode, Token{})
}

func NewWithToken(mode Mode, token Token) (*Client, error) {
  if mode == SANDBOX || mode == PRODUCTION {
    return &Client{mode, token, Credentials{}}, nil
  } else {
    return nil, errors.New("Invalid mode")
  }
}

// Expects token and api client credentials in the config file
// so that the client can:
//  - execute an authenticated API method using thetoken
//  - refresh the token sending client_id, client_secret and refresh_token - TokenCredentials
//  https://developers.bitwire.co/api/v1/#refresh-token
func NewFromConfig(mode Mode, config Config) (*Client, error) {
  if mode == SANDBOX || mode == PRODUCTION {
    return &Client{mode, config.Token, config.Credentials}, nil
  } else {
    return nil, errors.New("Invalid mode")
  }
}

// Returns the token
func (c *Client) Token() Token {
  return c.token
}

// Returns a Sling http clients configured with the base URL path
func (c *Client) http() *sling.Sling {
  switch c.Mode {
  case SANDBOX:
    return sling.New().Base(sandboxBaseURL)
  default:
    return sling.New().Base(baseURL)
  }
}

// Refreshes the token if it expires
func checkToken(c *Client) error {
  if c.token == (Token{}) {
    return errors.New("Missing auth token")
  }
  now := time.Now().Unix()
  if now >= c.token.ValidUntil-30 {
    _, err := c.RefreshToken()
    if err != nil {
      return err
    }
  }
  return nil
}

// General function for calling API method
// - sets auth headers
// - refreshes the token if necessary and parses error responses
func callApi(method Method, path string, params interface{}, c *Client, auth bool, res interface{}) error {
  var req *sling.Sling
  errorRes := new(ErrorRes)
  switch method {
  case POST:
    fallthrough
  case JSON_POST:
    req = c.http().Post(path)
  case DELETE:
    req = c.http().Delete(path)
  default:
    req = c.http().Get(path)
  }
  if auth {
    err := checkToken(c)
    if err != nil {
      return err
    }
    req.Set("Authorization", "Bearer "+c.token.AccessToken)
  }
  if params != nil {
    switch method {
    case JSON_POST:
      req = req.BodyJSON(params)
    case POST:
      req = req.BodyForm(params)
    default:
      req.QueryStruct(params)
    }

  }

  _, httpErr := req.Receive(res, errorRes)
  if httpErr != nil {
    return httpErr
  } else if *errorRes != (ErrorRes{}) {
    return errors.New(errorRes.ErrorType + ": " + errorRes.Message)
  } else {
    return nil
  }
}

func (c *Client) GetAllRates() (AllRates, error) {
  ratesRes := new(AllRatesRes)
  err := callApi(GET, "rates", nil, c, false, ratesRes)
  if err != nil {
    return AllRates{}, err
  } else {
    return ratesRes.Rates, nil
  }
}

func (c *Client) GetFxRates() (Rates, error) {
  ratesRes := new(FxRatesRes)
  err := callApi(GET, "rates/fx", nil, c, false, ratesRes)
  if err != nil {
    return nil, err
  } else {
    return ratesRes.Rates, nil
  }
}

func (c *Client) GetBtcRates() (Rates, error) {
  ratesRes := new(BtcRatesRes)
  err := callApi(GET, "rates/btc", nil, c, false, ratesRes)
  if err != nil {
    return nil, err
  } else {
    return ratesRes.Rates, nil
  }
}

func (c *Client) GetBanks() ([]Bank, error) {
  banksRes := new(BanksRes)
  err := callApi(GET, "banks", nil, c, false, banksRes)
  if err != nil {
    return nil, err
  } else {
    return banksRes.Banks, nil
  }
}

func (c *Client) GetRecipients() ([]Recipient, error) {
  recipientsRes := new(RecipientsRes)
  err := callApi(GET, "recipients", nil, c, true, recipientsRes)
  if err != nil {
    return nil, err
  } else {
    return recipientsRes.Recipients, nil
  }
}

func (c *Client) GetTransfers() ([]Transfer, error) {
  transfersRes := new(TransfersRes)
  err := callApi(GET, "transfers", nil, c, true, transfersRes)
  if err != nil {
    return nil, err
  } else {
    return transfersRes.Transfers, nil
  }
}

func (c *Client) GetTransfer(id string) (Transfer, error) {
  transferRes := new(TransferRes)
  err := callApi(GET, "transfers/"+id, nil, c, true, transferRes)
  if err != nil {
    return Transfer{}, err
  } else {
    return transferRes.Transfer, nil
  }
}

func (c *Client) CreateTransfer(transfer CreateTransfer) (Transfer, error) {
  transferRes := new(TransferRes)
  err := callApi(JSON_POST, "transfers", transfer, c, true, transferRes)
  if err != nil {
    return Transfer{}, err
  } else {
    return transferRes.Transfer, nil
  }
}

func (c *Client) CancelTransfer(id string) (Transfer, error) {
  transferRes := new(TransferRes)
  err := callApi(DELETE, "transfers/"+id, nil, c, true, transferRes)
  if err != nil {
    return Transfer{}, err
  } else {
    return transferRes.Transfer, nil
  }
}

func (c *Client) GetLimits() (Limits, error) {
  limitsRes := new(LimitsRes)
  err := callApi(GET, "users/limits", nil, c, true, limitsRes)
  if err != nil {
    return Limits{}, err
  } else {
    return limitsRes.Limits, nil
  }
}

// Calls direct auth method with username and password
// https://developers.bitwire.co/api/v1/#direct-authentication
func getToken(c *Client, credentials LoginCredentials) (Token, error) {
  tokenRes := new(TokenRes)
  err := callApi(POST, "oauth/tokens", credentials, c, false, tokenRes)
  if err != nil {
    return Token{}, err
  } else {
    token := tokenRes.Token
    token.ValidUntil = int64(token.ExpiresIn) + time.Now().Unix()
    return token, nil
  }
}

func (c *Client) TokenAuthenticate(credentials LoginCredentials, token Token) (Token, error) {
  return c.Authenticate(credentials)
}

// https://developers.bitwire.co/api/v1/#refresh-token
func refreshToken(c *Client, credentials TokenCredentials) (Token, error) {
  tokenRes := new(TokenRes)
  err := callApi(POST, "oauth/tokens", credentials, c, false, tokenRes)
  if err != nil {
    return Token{}, err
  } else {
    token := tokenRes.Token
    token.ValidUntil = int64(token.ExpiresIn) + time.Now().Unix()
    return token, nil
  }
}

func (c *Client) RefreshToken() (Token, error) {
  creds := TokenCredentials{c.credentials, c.token.RefreshToken}
  token, err := refreshToken(c, creds)
  if err == nil {
    c.token = token
  }
  return token, err
}

func (c *Client) Authenticate(credentials LoginCredentials) (Token, error) {
  token, err := getToken(c, credentials)
  if err != nil {
    return Token{}, err
  } else {
    c.credentials = Credentials{credentials.ClientId, credentials.ClientSecret, "refresh_token"}
    c.token = token
    return token, nil
  }
}
