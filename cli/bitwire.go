package main

import (
  "bufio"
  "encoding/base64"
  "encoding/json"
  "fmt"
  "github.com/dworznik/bitwire"
  "github.com/dworznik/cli"
  "io/ioutil"
  "os"
  "path/filepath"
  "strings"
)

const (
  ConfDir         = ".bitwire"
  ConfPath        = ConfDir + "/" + "production.json"
  SandboxConfPath = ConfDir + "/" + "sandbox.json"
)

type format string

const (
  TEXT format = "text"
  JSON format = "json"
  CSV  format = "csv"
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

func config(mode bitwire.Mode) (bitwire.Config, error) {
  fmt.Printf("Configuring bitwire in %s mode\n", mode)
  reader := bufio.NewReader(os.Stdin)
  fmt.Print("Username: ")
  username, _ := readStdin(reader)
  fmt.Print("Password: ")
  password, _ := readStdin(reader)
  fmt.Print("Client ID: ")
  clientId, _ := readStdin(reader)
  fmt.Print("Client secret: ")
  clientSecret, _ := readStdin(reader)
  conf := bitwire.Config{mode, username, password, clientId, clientSecret}
  return conf, nil
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
      pass, err := base64.StdEncoding.DecodeString(config.Password)
      if err != nil {
        return bitwire.Config{}, err
      } else {
        config.Password = string(pass)
        return config, nil
      }
    }
  }
}

func writeConfig(config bitwire.Config, mode bitwire.Mode) error {
  pass := base64.StdEncoding.EncodeToString([]byte(config.Password))
  config.Password = pass
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
  file, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE, 0666)
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

func printOut(v interface{}, format string) error {
  if format == "json" {
    output, err := formatJson(v)
    if err != nil {
      return cli.NewExitError(err.Error(), 10)
    } else {
      fmt.Println(output)
    }
  } else {
    return cli.NewExitError("Unknow format", 10)
  }
  return nil
}

func main() {
  authCommands := map[string]bool{"transfers": true, "limits": true, "recipients": true}
  sandbox := false
  mode := bitwire.PRODUCTION
  var format string

  var confErr error
  var conf bitwire.Config

  app := cli.NewApp()
  app.Name = "bitwire"
  app.Version = "0.0.1"
  app.Usage = "Bitwire command line interface"
  app.Flags = []cli.Flag{
    cli.BoolFlag{
      Name:        "sandbox, s",
      Usage:       "run in sandbox mode",
      Destination: &sandbox,
    },
    cli.StringFlag{
      Name:        "format, f",
      Usage:       "Output format: json, text",
      Value:       "json",
      Destination: &format,
    },
  }

  client := func(cmd string) (*bitwire.Client, error) {
    if authCommands[cmd] {
      if conf != (bitwire.Config{}) {
        client, err := bitwire.New(mode)
        if err != nil {
          return nil, cli.NewExitError(err.Error(), 1)
        } else {
          ok, err := client.Authenticate(bitwire.Credentials{Config: conf, GrantType: "password"})
          if err != nil {
            return nil, cli.NewExitError(err.Error(), 1)
          } else if ok {
            return client, nil
          } else {
            return nil, cli.NewExitError("Authentication failed", 1)
          }
        }
      } else {
        if confErr != nil {
          return nil, cli.NewExitError(confErr.Error(), 1)
        } else {
          return nil, cli.NewExitError("API authorization error", 1)
        }
      }
    } else {
      client, err := bitwire.New(mode)
      if err != nil {
        return nil, cli.NewExitError(err.Error(), 1)
      } else {
        return client, nil
      }
    }
  }

  app.Before = func(c *cli.Context) error {
    if sandbox {
      mode = bitwire.SANDBOX
      fmt.Println("Running in sandbox mode")
    } else {
      fmt.Println("Running in production mode")
    }
    conf, confErr = readConfig(mode)
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
        conf, err := config(mode)
        if err != nil {
          return cli.NewExitError(err.Error(), 1)
        } else {
          return writeConfig(conf, mode)
        }
      },
    },
    {
      Name:  "rates",
      Usage: "list current rates",
      Action: func(c *cli.Context) error {
        client, err := client(c.Command.Name)
        if err != nil {
          return err
        } else {
          rates, err := client.GetAllRates()
          if err != nil {
            return err
          } else {
            printOut(rates, format)
            return nil
          }
        }
      },
    },
    {
      Name:  "banks",
      Usage: "list banks",
      Action: func(c *cli.Context) error {
        client, err := client(c.Command.Name)
        if err != nil {
          return err
        } else {
          banks, err := client.GetBanks()
          if err != nil {
            return err
          } else {
            printOut(banks, format)
            return nil
          }
        }
      },
    },
    {
      Name:  "recipients",
      Usage: "list recipients",
      Action: func(c *cli.Context) error {
        client, err := client(c.Command.Name)
        if err != nil {
          return err
        } else {
          recipients, err := client.GetRecipients()
          if err != nil {
            return err
          } else {
            printOut(recipients, format)
            return nil
          }
        }
      },
    },
    {
      Name:  "transfers",
      Usage: "list transfers",
      Action: func(c *cli.Context) error {
        client, err := client(c.Command.Name)
        if err != nil {
          return err
        } else {
          txs, err := client.GetTransfers()
          if err != nil {
            return err
          } else {
            printOut(txs, format)
            return nil
          }
        }
      },
    },
    {
      Name:  "limits",
      Usage: "list limits",
      Action: func(c *cli.Context) error {
        client, err := client(c.Command.Name)
        if err != nil {
          return err
        } else {
          limits, err := client.GetLimits()
          if err != nil {
            return err
          } else {
            printOut(limits, format)
            return nil
          }
        }
      },
    },
  }
  app.Run(os.Args)
}
