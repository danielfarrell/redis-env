package main

import (
  "flag"
  "fmt"
  "github.com/hoisie/redis"
  "os"
  "os/exec"
  "strings"
  "strconv"
  "syscall"
)

func printVersion() {
  fmt.Println(fmt.Sprintf("%s 0.0.1", os.Args[0]))
  os.Exit(0)
}

func listConfig(client redis.Client, key string) {
  value := map[string][]byte{}
  err   := client.Hgetall(key, value)

  if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }

  for key, value := range value {
    fmt.Println(fmt.Sprintf("%s=%s", key, value))
  }
}

func addConfig(client redis.Client, key string, nameAndValue string) {
  parts := strings.Split(nameAndValue, "=")
  name := parts[0]
  value := strings.Join(parts[1:len(parts)], "=")
  _, err := client.Hset(key, name, []byte(value))

  if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }
}

func removeConfig(client redis.Client, key string, name string) {
  _, err := client.Hdel(key, name)

  if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }
}

func run(client redis.Client, key string) {
  value := map[string][]byte{}
  err   := client.Hgetall(key, value)

  if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(111)
  }

  envs := os.Environ()

  for key, value := range value {
    envs = append(envs, fmt.Sprintf("%s=%s", key, value))
  }

  cmd := exec.Command(flag.Args()[0], flag.Args()[1:]...)
  cmd.Env = envs
  cmd.Stderr = os.Stderr
  cmd.Stdout = os.Stdout
  cmd.Stdin = os.Stdin
  err = cmd.Run()
  if cmd.Process == nil {
    fmt.Fprintf(os.Stderr, "redis-env: %s\n", err)
    os.Exit(1)
  }
  os.Exit(cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus())
}

func main() {
  version := flag.Bool("version", false, "Print version and exit")
  list := flag.Bool("list", false, "List config vars")
  remove := flag.String("remove", "", "Config var to remove")
  add := flag.String("add", "", "Config var to add")
  flag.Parse()

  var key = os.Getenv("REDISENV_KEY")
  if key == "" {
    key = "default"
  }
  var host = os.Getenv("REDISENV_HOST")
  if host == "" {
    host = "127.0.0.1:6379"
  }
  var udb = os.Getenv("REDISENV_DB")
  var db int
  if udb != "" {
    db, _ = strconv.Atoi(udb)
  }

  var client redis.Client
  client.Addr = host
  client.Db = db

  if *version {
    printVersion()
  } else if *list {
    listConfig(client, key)
  } else if len(*remove) > 0 {
    removeConfig(client, key, *remove)
  } else if len(*add) > 0 {
    addConfig(client, key, *add)
  } else {
    run(client, key)
  }
}
