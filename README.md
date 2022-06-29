# go xmpp server
- First attempt at writing a minimal xmpp server to learn golang.
- Used https://github.com/shapeshed/golang-book-examples/blob/master/hour26/example05.go as a template
- Provides only ANONYMOUS SASL, STARTTLS & Resource Binding Stream Features.
- A bit hodge podge.
- Runs on localhost only.

### Run example app
```
cd _example
./makeCert.sh
go mod init example
go mod tidy
go run main.go
```

### Test with xmpp client
```
git clone https://github.com/jrlee89/go-xmpp-example.git
```
Use tmux and run two instances of the example program.
```
go run example.go -server=localhost -notls=true -debug=true
```



