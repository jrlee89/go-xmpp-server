# go xmpp server
- First attempt at writing a minimal xmpp server to learn golang.
- Inspired by https://github.com/mattn/go-xmpp/
- Provides only ANONYMOUS SASL, STARTTLS & Server-Generated Resource Binding Stream Features.
- Runs on localhost only.

### Run the server
```
./makeCert.sh
go run *.go
```

### Test with xmpp client
```
git clone https://github.com/jrlee89/go-xmpp-example.git
```
Use tmux and run two instances of the example program.
```
go run example.go -server=localhost -notls=true -debug=true
```



