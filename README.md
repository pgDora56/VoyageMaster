# VoyageMaster

Monitors the Discord server and notifies you of channel comings and goings.

## Requirement / Environment

Refer to `go.mod`


## Build

in Linux

```bash
make                // Build
./voyagemaster      // Run
```

## Setting

All settings are written in `config.toml`.

For example:
```toml
# Bot 1
[[bot]]
  name = "Mocho"
  token = "NI9ud9340jofdivjpoisadjfgopiwrhpg9uhadpf"
  deletetime = 10
  join = "{user}が{channel}に参加しました！"
  move = "{user}が{before}から{after}に移動しました！"
  leave = "{user}が{channel}から退出しました！"

  [[bot.targets]]
    category = "984756014843541"
    sendto = "7821463519457945"

# Bot 2
[[bot]]
  name = "Tenka"
  token = "NZewtfhsdiopuvf34ourhfgv9087erhny5bfge89h34g"
  deletetime = 30
  join = "{user}: {channel}に参加"
  move = "{user}: {before} -> {after} へ移動"
  leave = "{user}: {channel}から退出"

  [[bot.targets]]
    category = "32948701984305"
    sendto = "89034234985719386"
```


### Interactive setting

```bash
./voyagemaster setting
```

#### Execution example

```bash
Bot name:Tenka
Discord Bot Token:NZewtfhsdiopuvf34ourhfgv9087erhny5bfge89h34g
Target category id(contain voice channels to be monitored):32948701984305
Text channel id(to send notifications to):89034234985719386
Do you want to add other targets? Please enter `yes` or `no`.
yes/no:no
Time to delete notification(natural number only(>0, integer)):30
Please enter Join notification message template(`{user}`->username, `{channel}`->channel to join)
Join:{user}: {channel}に参加
Please enter Move notification message template(`{user}`->username, `{before}`->channel to leave, `{after}`->channel to join)
Move:{user}: {before} -> {after} へ移動
Please enter Leave notification message template(`{user}`->username, `{channel}`->channel to leave)
Leave:{user}: {channel}から退出
Do you want to add other setting? Please enter `yes` or `no`.
yes/no:no
```

## LICENSE

[MIT](LICENSE)