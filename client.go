package bitwire

import (
  "errors"
  "fmt"
  "github.com/dghubble/sling"
  "net/http"
)

const baseURL = "https://www.bitwire.co/api/v1/"
const sandboxBaseURL = "https://sandbox.bitwire.co/api/v1/"

type Res struct {
  Code int `json:"code"`
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

func New(mode Mode) (*Client, error) {
  if mode == SANDBOX || mode == PRODUCTION {
    return &Client{mode, Token{}}, nil
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

func (c *Client) GetAllRates() (AllRates, error) {
  ratesRes := new(AllRatesRes)
  fmt.Println("Fetching all rates")
  _, err := c.http().Get("rates").Receive(ratesRes, nil)
  if err != nil {
    return AllRates{}, err
  } else {
    return ratesRes.Rates, nil
  }
}

func (c *Client) GetFxRates() (Rates, error) {
  ratesRes := new(FxRatesRes)
  fmt.Println("Fetching FX rates")
  _, err := c.http().Get("rates/fx").Receive(ratesRes, nil)
  if err != nil {
    return nil, err
  } else {
    return ratesRes.Rates, nil
  }
}

func (c *Client) GetBtcRates() (Rates, error) {
  ratesRes := new(BtcRatesRes)
  fmt.Println("Fetching BTC rates")
  _, err := c.http().Get("rates/btc").Receive(ratesRes, nil)
  if err != nil {
    return nil, err
  } else {
    return ratesRes.Rates, nil
  }
}

func (c *Client) GetBanks() ([]Bank, error) {
  banksRes := new(BanksRes)
  fmt.Println("Fetching banks")
  _, err := c.http().Get("banks").Receive(banksRes, nil)
  if err != nil {
    return nil, err
  } else {
    return banksRes.Banks, nil
  }
}

func (c *Client) GetRecipients() ([]Recipient, error) {
  recipientsRes := new(RecipientsRes)
  req := c.http().Get("recipients")
  if c.token != (Token{}) {
    req.Set("Authorization", "Bearer "+c.token.AccessToken)
  }
  res, err := req.Body(nil).Receive(recipientsRes, nil)
  if apiErr := check(res, err); apiErr != nil {
    return nil, apiErr
  } else {
    return recipientsRes.Recipients, nil
  }
}

func (c *Client) GetTransfers() ([]Transfer, error) {
  transfersRes := new(TransfersRes)
  req := c.http().Get("transfers")
  if c.token != (Token{}) {
    req.Set("Authorization", "Bearer "+c.token.AccessToken)
  }
  res, err := req.Body(nil).Receive(transfersRes, nil)
  if apiErr := check(res, err); apiErr != nil {
    return nil, apiErr
  } else {
    return transfersRes.Transfers, nil
  }
}

func (c *Client) GetLimits() (Limits, error) {
  limitsRes := new(LimitsRes)
  req := c.http().Get("users/limits")
  if c.token != (Token{}) {
    req.Set("Authorization", "Bearer "+c.token.AccessToken)
  }
  res, err := req.Body(nil).Receive(limitsRes, nil)
  if apiErr := check(res, err); apiErr != nil {
    return Limits{}, apiErr
  } else {
    return limitsRes.Limits, nil
  }
}

func (c *Client) GetToken(credentials Credentials) (Token, error) {
  tokenRes := new(TokenRes)
  res, err := c.http().Post("oauth/tokens").BodyForm(credentials).Receive(tokenRes, nil)
  if apiErr := check(res, err); apiErr != nil {
    return Token{}, apiErr
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

func check(res *http.Response, err error) error {
  if err != nil {
    return err
  } else if res.StatusCode != 200 {
    return errors.New(fmt.Sprintf("HTTP status: %d", res.StatusCode))
  } else {
    return nil
  }
}
