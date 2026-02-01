# Fife
A simple go server for sending WOL packetsS. 

## Example
```yaml
# ./.fife.yaml
bindHost: ":8080"
services:
  foo:
    to: foo.com
    wol:
      mac: aa:bb:cc:dd:ee
```
```
fife wol-server
```

## TODO
- fix github action (docker?)
- add /etc/fife/fife.yaml as default config location
