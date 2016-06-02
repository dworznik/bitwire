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
  Currency  string            `json:"currecny"`
  Status    string            `json:"status"`
  Date      string            `json:"date"`
  BTC       BTC               `json:"btc"`
  Recipient TransferRecipient `json:"recipient"`
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

type Config struct {
  Mode         Mode   `json:"mode"`
  Username     string `json:"username" url:"username"`
  Password     string `json:"password" url:"password"`
  ClientId     string `json:"client_id" url:"client_id"`
  ClientSecret string `json:"client_secret" url:"client_secret"`
}

type Credentials struct {
  Config
  GrantType string `url:"grant_type"`
}

type Client struct {
  Mode  Mode
  token Token
}

type Method string

const (
  GET  Method = "GET"
  POST Method = "POST"
)

func New(mode Mode) (*Client, error) {
  return NewWithToken(mode, Token{})
}

func NewWithToken(mode Mode, token Token) (*Client, error) {
  if mode == SANDBOX || mode == PRODUCTION {
    return &Client{mode, token}, nil
  } else {
    return nil, errors.New("Invalid mode")
  }
}

func (c *Client) http() *sling.Sling {
  switch c.Mode {
  case SANDBOX:
    return sling.New().Base(sandboxBaseURL)
  default:
    return sling.New().Base(baseURL)
  }
}

func callApi(method Method, path string, params interface{}, c *Client, auth bool, res interface{}) error {
  var req *sling.Sling
  errorRes := new(ErrorRes)
  switch method {
  case POST:
    req = c.http().Post(path)
  default:
    req = c.http().Get(path)
  }
  if auth {
    req.Set("Authorization", "Bearer "+c.token.AccessToken)
  }
  if params != nil {
    req = req.BodyForm(params)
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

func (c *Client) GetLimits() (Limits, error) {
  limitsRes := new(LimitsRes)
  err := callApi(GET, "users/limits", nil, c, true, limitsRes)
  if err != nil {
    return Limits{}, err
  } else {
    return limitsRes.Limits, nil
  }
}

func (c *Client) GetToken(credentials Credentials) (Token, error) {
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

func (c *Client) TokenAuthenticate(credentials Credentials, token Token) (bool, error) {
  return c.Authenticate(credentials)
}

func (c *Client) RefreshToken(credentials Credentials, token Token) (Token, error) {
  tokenRes := new(TokenRes)
  req := struct {
    ClientId     string `url:"client_id"`
    ClientSecret string `url:"client_secret"`
    RefreshToken string `url:"refresh_token"`
    GrantType    string `url:"grant_type"`
  }{credentials.ClientId, credentials.ClientSecret, token.RefreshToken, "refresh_token"}
  err := callApi(POST, "oauth/tokens", req, c, false, tokenRes)
  if err != nil {
    return Token{}, err
  } else {
    return tokenRes.Token, nil
  }
}

func (c *Client) Authenticate(credentials Credentials) (bool, error) {
  token, err := c.GetToken(credentials)
  if err != nil {
    return false, err
  } else {
    c.token = token
    return true, nil
  }
}
