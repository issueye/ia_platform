# Per-Library Examples

这个目录按“每个原生库一个最小示例”的方式组织，方便对照 wiki 文档逐个试跑。

## Network

- `http.ia`
- `websocket.ia`
- `sse.ia`
- `express.ia`
- `ipc.ia`
- `iax.ia`
- `socket.ia`
- `net.ia`

## System

- `fs.ia`
- `path.ia`
- `os.ia`
- `process.ia`
- `signal.ia`
- `exec.ia`
- `time.ia`
- `timer.ia`
- `pool.ia`
- `asset.ia`

## Data And Utils

- `math.ia`
- `string.ia`
- `array.ia`
- `sort.ia`
- `set.ia`
- `strconv.ia`
- `rand.ia`
- `log.ia`
- `promise.ia`
- `agent-sdk.ia`
- `json.ia`
- `yaml.ia`
- `toml.ia`
- `xml.ia`
- `csv.ia`
- `encoding.ia`
- `hex.ia`
- `bytes.ia`
- `mime.ia`
- `crypto.ia`
- `hash.ia`
- `hmac.ia`
- `uuid.ia`
- `regexp.ia`
- `url.ia`
- `db.ia`
- `orm.ia`
- `compress.ia`

## Batch Check

在 `ialang` 子项目根目录执行：

```bash
Get-ChildItem .\examples\wiki_examples\libraries\*.ia | ForEach-Object { go run ./cmd/ialang check $_.FullName }
```
