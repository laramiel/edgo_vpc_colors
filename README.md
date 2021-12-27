# edgo_vpc_colors

Elite Dangerous journal watcher to set Virpil VPC colors based on in-game events.

Somewhat related to https://github.com/charliefoxtwo/ViLA
or https://github.com/Painter602/EDLogReader

Has no configuration file and is based on my older ed journal parser.

## How to install

```
go get github.com/fsnotify/fsnotify

GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui"
```

