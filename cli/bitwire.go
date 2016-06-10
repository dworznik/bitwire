package main

import (
  "bufio"
  "encoding/json"
  "fmt"
  "github.com/dworznik/bitwire"
  "github.com/dworznik/cli"
  "github.com/olekukonko/tablewriter"
  "io/ioutil"
  "os"
  "path/filepath"
  "strings"
)

func printfErr(format string, v ...interface{}) (n int, err error) {
  return fmt.Fprintf(os.Stderr, format, v...)
}

const (
  ConfDir         = ".bitwire"
  ConfPath        = ConfDir + "/" + "production.json"
  SandboxConfPath = ConfDir + "/" + "sandbox.json"
)

func configDir() string {
  return filepath.FromSlash(os.Getenv("HOME") + "/" + ConfDir)
}

func configPath(mode bitwire.Mode) string {
  switch mode {
  case bitwire.SANDBOX:
    return filepath.FromSlash(os.Getenv("HOME") + "/" + SandboxConfPath)
  case bitwire.PRODUCTION:
    return filepath.FromSlash(os.Getenv("HOME") + "/" + ConfPath)
  default:
    panic("Missing mode")
  }
}

func readStdin(reader *bufio.Reader) (string, error) {
  val, err := reader.ReadString('\n')
  if err != nil {
    return val, err
  } else {
    return strings.TrimRight(val, "\n"), nil
  }
}

func config(mode bitwire.Mode) (bitwire.Config, bitwire.LoginCredentials, error) {
  printfErr("Configuring bitwire in %s mode\n", mode)
  reader := bufio.NewReader(os.Stdin)
  fmt.Print("Username: ")
  username, _ := readStdin(reader)
  fmt.Print("Password: ")
  password, _ := readStdin(reader)
  fmt.Print("Client ID: ")
  clientId, _ := readStdin(reader)
  fmt.Print("Client secret: ")
  clientSecret, _ := readStdin(reader)
  tokenCreds := bitwire.Credentials{clientId, clientSecret, "refresh_token"}
  passwordCreds := bitwire.Credentials{clientId, clientSecret, "password"}
  conf := bitwire.Config{tokenCreds, bitwire.Token{}}
  login := bitwire.LoginCredentials{passwordCreds, username, password}
  return conf, login, nil
}

func readConfig(mode bitwire.Mode) (bitwire.Config, error) {
  data, err := ioutil.ReadFile(configPath(mode))
  if err != nil {
    return bitwire.Config{}, err
  } else {
    config := bitwire.Config{}
    err := json.Unmarshal(data, &config)
    if err != nil {
      return config, err
    } else {
      return config, nil
    }
  }
}

func writeConfig(config bitwire.Config, mode bitwire.Mode) error {
  configDir := configDir()
  configPath := configPath(mode)
  err := os.Mkdir(configDir, 0777)
  if err != nil {
    if _, ok := err.(*os.PathError); ok {
      // Config dir already exists
    } else {
      return cli.NewExitError(err.Error(), 1)
    }
  }
  file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
  if err != nil {
    return cli.NewExitError(err.Error(), 1)
  } else {
    defer file.Close()
    str, err := formatJson(config)
    if err != nil {
      return cli.NewExitError(err.Error(), 1)
    } else {
      file.WriteString(str)
      return nil
    }
  }
}

func formatJson(v interface{}) (string, error) {
  b, err := json.MarshalIndent(v, "", "  ")
  if err != nil {
    return "", err
  } else {
    return string(b), nil
  }
}

var tableTransferHeader = []string{"ID", "Recipient", "Sent (BTC)", "Received", "Date", "Status"}

func tableTransferData(transfer bitwire.Transfer) []string {
  return []string{transfer.Id,
    transfer.Recipient.Name,
    fmt.Sprintf("%s %s", transfer.Amount, transfer.Currency),
    fmt.Sprintf("%s %s", transfer.Recipient.Amount, transfer.Recipient.Currency),
    transfer.Date, transfer.Status}
}

var tableRecipientHeader = []string{"ID", "Name", "Email", "Bank", "Account"}

func tableRecipientData(recipient bitwire.Recipient) []string {
  return []string{fmt.Sprintf("%d", recipient.Id), recipient.Name, recipient.Email, recipient.Bank.DisplayName, recipient.Bank.AccountNumber}
}

var tableBankHeader = []string{"ID", "Number", "Name"}

func tableBankData(bank bitwire.Bank) []string {
  return []string{fmt.Sprintf("%d", bank.Id), bank.Number, bank.Name}
}

func tableLimitData(limit bitwire.Limits) []string {
  return nil
}

var tableRatesHeader = []string{"", "Rate"}

var tableLimitsHeader = []string{"Limit", "Value (BTW)"}

var tableTransferLimitsHeader = []string{"Limit", "Value"}

func printOut(obj interface{}, json bool) error {
  if json {
    output, err := formatJson(obj)
    if err != nil {
      return cli.NewExitError(err.Error(), 10)
    } else {
      fmt.Println(output)
    }
  } else {
    table := tablewriter.NewWriter(os.Stdout)
    switch v := obj.(type) {
    case []bitwire.Transfer:
      table.SetHeader(tableTransferHeader)
      for i := range v {
        table.Append(tableTransferData(v[i]))
      }
    case []bitwire.Recipient:
      table.SetHeader(tableRecipientHeader)
      for i := range v {
        table.Append(tableRecipientData(v[i]))
      }
    case []bitwire.Bank:
      table.SetHeader(tableBankHeader)
      for i := range v {
        table.Append(tableBankData(v[i]))
      }
    case bitwire.AllRates:
      table.SetHeader(tableRatesHeader)
      for k, v := range v.BTC {
        table.Append([]string{k, v})
      }
      table.Append([]string{"", ""})
      for k, v := range v.FX {
        table.Append([]string{k, v})
      }
    case bitwire.Limits:
      table.SetHeader(tableLimitsHeader)
      table.Append([]string{"Daily used", v.KRW.Daily.Used})
      table.Append([]string{"Daily left", v.KRW.Daily.Left})
      table.Append([]string{"Daily limit", v.KRW.Daily.Limit})
      table.Append([]string{"Weekly used", v.KRW.Weekly.Used})
      table.Append([]string{"Weekly left", v.KRW.Weekly.Left})
      table.Append([]string{"Weekly limit", v.KRW.Weekly.Limit})
      table.Render()

      table = tablewriter.NewWriter(os.Stdout)
      table.SetHeader(tableTransferLimitsHeader)
      table.Append([]string{"Pending transfers used", fmt.Sprintf("%d", v.Transfers.Pending.Total.Used)})
      table.Append([]string{"Pending transfers limit", fmt.Sprintf("%d", v.Transfers.Pending.Total.Limit)})
      table.Append([]string{"Daily transfers used", fmt.Sprintf("%d", v.Transfers.Completed.Daily.Used)})
      table.Append([]string{"Daily transfers limit", fmt.Sprintf("%d", v.Transfers.Completed.Daily.Limit)})
    }

    table.Render()
  }
  return nil
}

func main() {
  var exit error
  defer func() {
    if exit != nil {
      printfErr("%s\n", exit)
      if exit.Error() == "Unauthorized: Token expired." {
        printfErr("API token could not been refreshed. Run bitwire config again\n")
      }
      os.Exit(1)
    }
  }()

  authCommands := map[string]bool{"transfers": true, "limits": true, "recipients": true}
  sandbox := false
  mode := bitwire.PRODUCTION
  var json = false

  var confErr error
  var conf bitwire.Config    // Set in app.Before()
  var client *bitwire.Client // Set in newClient()

  app := cli.NewApp()
  app.Name = "bitwire"
  app.Version = "0.0.2"
  app.Usage = "Bitwire command line interface"
  app.Flags = []cli.Flag{
    cli.BoolFlag{
      Name:        "sandbox, s",
      Usage:       "run in sandbox mode",
      Destination: &sandbox,
    },
    cli.BoolFlag{
      Name:        "json, j",
      Usage:       "print out JSON",
      Destination: &json,
    },
  }

  // newClient creates a new bitwire client for running a client
  // Returns an error if the command requires authentication and it cannot read credentials from the config file
  newClient := func(cmd string) (*bitwire.Client, error) {
    if authCommands[cmd] {
      if conf != (bitwire.Config{}) {
        c, err := bitwire.NewFromConfig(mode, conf)
        if err != nil {
          return nil, cli.NewExitError(err.Error(), 1)
        } else {
          client = c
          return client, nil
        }
      } else {
        if confErr != nil {
          return nil, cli.NewExitError(confErr.Error(), 1)
        } else {
          return nil, cli.NewExitError("Configuration error", 1)
        }
      }
    } else {
      c, err := bitwire.New(mode)
      if err != nil {
        return nil, cli.NewExitError(err.Error(), 1)
      } else {
        client = c
        return client, nil
      }
    }
  }

  app.Before = func(c *cli.Context) error { // Read config from the file before running a command
    if sandbox {
      mode = bitwire.SANDBOX
      printfErr("Running in sandbox mode\n")
    } else {
      printfErr("Running in production mode\n")
    }
    conf, confErr = readConfig(mode)
    return nil
  }

  app.After = func(c *cli.Context) error {
    if client != nil {
      if client.Token().AccessToken != "" && conf.Token.AccessToken != client.Token().AccessToken { // Update token in the config file
        conf = bitwire.Config{bitwire.Credentials{conf.ClientId, conf.ClientSecret, conf.GrantType}, client.Token()}
        return writeConfig(conf, mode)
      }
    }
    return nil
  }

  app.OnUsageError = func(context *cli.Context, err error, isSubcommand bool) error {
    return nil
  }

  app.Action = func(c *cli.Context) error {
    cli.ShowAppHelp(c)
    return nil
  }

  app.CommandNotFound = func(c *cli.Context, cmd string) {
    fmt.Println("Unrecognized command: ", cmd)
    cli.ShowAppHelp(c)
  }

  app.Commands = []cli.Command{
    {
      Name:  "config",
      Usage: "configure Bitwire API access",
      Action: func(c *cli.Context) error {
        client, err := newClient(c.Command.Name)
        if exit = err; err != nil {
          return err
        }
        conf, login, err := config(mode)
        if exit = err; err != nil {
          return err
        }
        token, err := client.Authenticate(login)
        if exit = err; err != nil {
          return err
        } else {
          conf.Token = token
          defer printfErr("Configuration saved\n")
          return writeConfig(conf, mode)
        }
      },
    },
    {
      Name:  "rates",
      Usage: "list current rates",
      Action: func(c *cli.Context) error {
        client, err := newClient(c.Command.Name)
        if exit = err; err != nil {
          return err
        } else {
          rates, err := client.GetAllRates()
          if exit = err; err != nil {
            return err
          } else {
            printOut(rates, json)
            return nil
          }
        }
      },
    },
    {
      Name:  "banks",
      Usage: "list banks",
      Action: func(c *cli.Context) error {
        client, err := newClient(c.Command.Name)
        if exit = err; err != nil {
          return err
        } else {
          banks, err := client.GetBanks()
          if exit = err; err != nil {
            return err
          } else {
            printOut(banks, json)
            return nil
          }
        }
      },
    },
    {
      Name:  "recipients",
      Usage: "list recipients",
      Action: func(c *cli.Context) error {
        client, err := newClient(c.Command.Name)
        if exit = err; err != nil {
          return err
        } else {
          recipients, err := client.GetRecipients()
          if exit = err; err != nil {
            return err
          } else {
            printOut(recipients, json)
            return nil
          }
        }
      },
    },
    {
      Name:  "transfers",
      Usage: "list transfers",
      Action: func(c *cli.Context) error {
        client, err := newClient(c.Command.Name)
        if exit = err; err != nil {
          return err
        } else {
          txs, err := client.GetTransfers()
          if exit = err; err != nil {
            return err
          } else {
            printOut(txs, json)
            return nil
          }
        }
      },
    },
    {
      Name:  "limits",
      Usage: "list limits",
      Action: func(c *cli.Context) error {
        client, err := newClient(c.Command.Name)
        if exit = err; err != nil {
          return err
        } else {
          limits, err := client.GetLimits()
          if exit = err; err != nil {
            return err
          } else {
            printOut(limits, json)
            return nil
          }
        }
      },
    },
  }
  app.Run(os.Args)
}
