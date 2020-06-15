Install GOLANG (I used 1.13).

Next:
```
mkdir -p ~/go
export GOPATH="$HOME/go"
export PATH="$GOPATH/bin:$PATH"
```

Install "mage":
```
git clone https://github.com/magefile/mage
cd mage
go run bootstrap.go
```

Download this repo (path matters):
```
mkdir -p ~/go/src/github.com/FracKenA/
cd ~/go/src/github.com/FracKenA/
git clone https://github.com/emca-it/op5beat
```

Build binary:
```
mage build
```
