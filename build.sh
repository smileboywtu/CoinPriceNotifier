GOOS=linux GOARCH=amd64 go build -o CoinPriceNotify \
 -ldflags "-w -s -X main.version=0.0.1 -X main.email=smileboywtu@gmail.com -X main.author=smileboywtu" \
 *.go