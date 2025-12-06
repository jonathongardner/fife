# Fife
A simple go reverse proxy. 

## Example
```yaml
# ./.fife.yaml
bindHost: ":8080"
services:
  - name: foo.localhost
    host: http://192.168.1.2:80
  - name: bar.localhost
    host: http://192.168.1.3:3000
```
```
fife reverse-proxy
```

## TODO
- fix github action (docker?)
- add /etc/fife/fife.yaml as default config location
- add simple default page showing config
- add globing/path 