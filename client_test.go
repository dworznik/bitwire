package bitwire

import (
  "encoding/base64"
  "encoding/json"
  "fmt"
  "github.com/stretchr/testify/assert"
  "io/ioutil"
  "testing"
  "time"
)

func TestClient(t *testing.T) {
  client, err := New("test")
  assert.Nil(t, client)
  assert.NotNil(t, err)
  assert.Equal(t, err.Error(), "Invalid mode")
}

func TestAllRates(t *testing.T) {
  client, _ := New(SANDBOX)
  rates, err := client.GetAllRates()
  assert.Nil(t, err)
  assert.NotEmpty(t, rates)
  assert.NotEmpty(t, rates.BTC)
  assert.Contains(t, rates.BTC, "BTCKRW")
  assert.NotEmpty(t, rates.FX)
}

func TestBtcRates(t *testing.T) {
  client, _ := New(SANDBOX)
  rates, err := client.GetBtcRates()
  assert.Nil(t, err)
  assert.NotEmpty(t, rates)
  assert.Contains(t, rates, "BTCKRW")
}

func TestFxRates(t *testing.T) {
  client, _ := New(SANDBOX)
  rates, err := client.GetFxRates()
  assert.Nil(t, err)
  assert.NotEmpty(t, rates)
}

func TestBanks(t *testing.T) {
  client, _ := New(SANDBOX)
  banks, err := client.GetBanks()
  assert.Nil(t, err)
  assert.NotEmpty(t, banks)
}

func TestToken(t *testing.T) {
  client, _ := New(SANDBOX)
  config := readConfig()
  creds := Credentials{Config: config, GrantType: "password"}
  token, err := client.GetToken(creds)
  fmt.Println(token)
  assert.Nil(t, err)
  assert.NotNil(t, token)
  assert.NotNil(t, token.AccessToken)
}

func TestAuthenticate(t *testing.T) {
  client, _ := New(SANDBOX)
  config := readConfig()
  creds := Credentials{Config: config, GrantType: "password"}
  ok, err := client.Authenticate(creds)
  assert.Nil(t, err)
  assert.True(t, ok)
}

func TestTransfers(t *testing.T) {
  client, _ := New(SANDBOX)
  config := readConfig()
  creds := Credentials{Config: config, GrantType: "password"}
  client.Authenticate(creds)
  transfers, err := client.GetTransfers()
  assert.Nil(t, err)
  assert.NotEmpty(t, transfers)
}

func TestLimits(t *testing.T) {
  client, _ := New(SANDBOX)
  config := readConfig()
  creds := Credentials{Config: config, GrantType: "password"}
  client.Authenticate(creds)
  limits, err := client.GetLimits()
  assert.Nil(t, err)
  assert.NotEmpty(t, limits)
}

func TestLimitsAuthFailed(t *testing.T) {
  token := Token{"Bearer",
    "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VyIjo5MSwibGV2ZWwiOjEsImVtYWlsIjoiZHd1eml1QGJ6aXVtLm5ldCIsImp0aSI6IjBQR1kyOEdtaEE3cjBUR1RYb3AwRzBjb3RmemU2aTd0IiwiaWF0IjoxNDY0Njc5ODIzLCJleHAiOjE0NjQ2ODM0MjMsImlzcyI6Imh0dHBzOi8vd3d3LmJpdHdpcmUuY28vYXBpL3YxL29hdXRoIn0.NE9gjpcaQimsTjyaWQncmJ67c6rdzlvFlaR12lskgWw",
    "xxx",
    3600,
    time.Now().Unix() + 3600,
  }
  client, _ := NewWithToken(SANDBOX, token)
  _, err := client.GetLimits()
  assert.NotNil(t, err)
  assert.Equal(t, err.Error(), "Unauthorized: Invalid token.")
}

func TestRecipients(t *testing.T) {
  client, _ := New(SANDBOX)
  config := readConfig()
  creds := Credentials{Config: config, GrantType: "password"}
  client.Authenticate(creds)
  recipients, err := client.GetRecipients()
  assert.Nil(t, err)
  assert.NotEmpty(t, recipients)
}

func TestRefreshToken(t *testing.T) {
  client, _ := New(SANDBOX)
  config := readConfig()
  creds := Credentials{Config: config, GrantType: "password"}
  token, err := client.GetToken(creds)
  newToken, err := client.RefreshToken(creds, token)
  assert.Nil(t, err)
  assert.NotNil(t, newToken)
  assert.NotNil(t, newToken.AccessToken)
  assert.NotEqual(t, token.AccessToken, newToken.AccessToken)
}

func readConfig() Config {
  data, err := ioutil.ReadFile("./test_sandbox.conf")
  if err != nil {
    panic(err)
  } else {
    config := Config{}
    err := json.Unmarshal(data, &config)
    if err != nil {
      panic(err)
    } else {
      pass, err := base64.StdEncoding.DecodeString(config.Password)
      if err != nil {
        panic(err)
      } else {
        config.Password = string(pass)
        return config
      }
    }
  }
}
