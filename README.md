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

```
token = "TOKENTOKENTOKENTOKENTOKENTOKENTOKENTOKENTOKENTOKEN"
targets = [{category = "{category_id}", sendto = "{text_channel_id}"}]
deletetime = {Time to delete(sec)}

[templates]
join = "{user}が{channel}に参加しました！"
move = "{user}が{before}から{after}に移動しました！"
leave = "{user}が{channel}から退出しました！"
```


## LICENSE

[MIT](LICENSE)