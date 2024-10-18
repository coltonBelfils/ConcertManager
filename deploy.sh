go get;

if [ -f ~/go/bin/golangci-lint ]; then
  ~/go/bin/golangci-lint run --disable=typecheck ./... || return 1
else
  echo "golangci-lint not found"
fi

go test -v ./... | grep -v "no test files"

go build

chmod +rwx ./SSRTest

COOKIE_CHECK_SALT=dragonscreemhelpqueen PASSWORD_SALT=crownpleasevanilla ADMIN_PASSWORD=ThinBlueHighway67 ./SSRTest

scp