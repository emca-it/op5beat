Zainstalować GOLANG (ja używałem 1.13).

Następnie:
```
mkdir -p ~/go
export GOPATH="$HOME/go"
export PATH="$GOPATH/bin:$PATH"
```

Zainstalować "mage":
```
git clone https://github.com/magefile/mage
cd mage
go run bootstrap.go
```

Pobrać to repo (ścieżka ma znaczenie):
```
mkdir -p ~/go/src/github.com/FracKenA/
cd ~/go/src/github.com/FracKenA/
git clone https://git.emca.pl/7.X/op5beat.git
```

Zbudować binarkę:
```
mage build
```

PS. Ja to robiłem na Fedora 31 ale wydaje mi się, że ze względu na użycie GO nie powinno być problemu na żadnym systemie.
