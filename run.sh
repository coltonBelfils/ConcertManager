go get;

if [ -f ~/go/bin/golangci-lint ]; then
  echo "Running golangci-lint"
  ~/go/bin/golangci-lint run --disable=typecheck ./... || return 1
else
  echo "golangci-lint not found"
fi

echo "Running tests, if applicable"
go test -v ./... | grep -v "no test files"

echo "Building"
go build

chmod +rwx ./ConcertGetApp

config_file="./conf.txt"

COOKIE_CHECK_SALT=$(awk -F= '/^cookieCheckSalt/ {print $2}' "$config_file")
PASSWORD_SALT=$(awk -F= '/^passwordSalt/ {print $2}' "$config_file")
ADMIN_PASSWORD=$(awk -F= '/^adminPassword/ {print $2}' "$config_file")
APP_ROOT=$(awk -F= '/^appRoot/ {print $2}' "$config_file")

if [ -z "$COOKIE_CHECK_SALT" ]; then
  echo "Error: cookieCheckSalt is missing from the config file"
  exit 1
fi

if [ -z "$PASSWORD_SALT" ]; then
  echo "Error: passwordSalt is missing from the config file"
  exit 1
fi

if [ -z "$ADMIN_PASSWORD" ]; then
  echo "Error: adminPassword is missing from the config file"
  exit 1
fi

if [ -z "$APP_ROOT" ]; then
  echo "Error: appRoot is missing from the config file"
  exit 1
fi

echo "Running"
COOKIE_CHECK_SALT=$COOKIE_CHECK_SALT PASSWORD_SALT=$PASSWORD_SALT ADMIN_PASSWORD=$ADMIN_PASSWORD APP_ROOT=$APP_ROOT ./ConcertGetApp