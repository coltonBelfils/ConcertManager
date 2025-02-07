go get;

if [ -f ~/go/bin/golangci-lint ]; then
  echo "Running golangci-lint"
  ~/go/bin/golangci-lint run --disable=typecheck ./... || return 1
else
  echo "golangci-lint not found"
fi

echo "Running tests, if applicable"
go test -v ./... | grep -v "no test files"

rm ./ConcertGetApp

echo "Building"
go build -o ConcertGetApp

chmod +rwx ./ConcertGetApp

if [ "$(which ./yt-dlp)" = "./yt-dlp not found" ]; then
  curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o ./yt-dlp > /dev/null 2>&1
  chmod a+rx ./yt-dlp > /dev/null
  if [ "$(which ./yt-dlp)" = "./yt-dlp not found" ]; then
    echo "Attempted to download yt-dlp but failed."
    exit 1
  fi
else
  ./yt-dlp -U > /dev/null
fi

alias yt-dlp='./yt-dlp'

version=$(yt-dlp --version)

if [ $? -ne 0 ]; then
  echo "yt-dlp not found in appropriate directory. It should be there and should have already been downloaded in the script if it was not there."
  exit 1
fi

echo "yt-dlp version: $version"

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

# Code signing
codesign --force --deep --sign - ./ConcertGetApp > /dev/null 2>&1

echo "Running"
export COOKIE_CHECK_SALT PASSWORD_SALT ADMIN_PASSWORD APP_ROOT
./ConcertGetApp 2>&1

# COOKIE_CHECK_SALT=$COOKIE_CHECK_SALT PASSWORD_SALT=$PASSWORD_SALT ADMIN_PASSWORD=$ADMIN_PASSWORD APP_ROOT=$APP_ROOT ./ConcertGetApp 2>&1 | tee error.log # doesn't work
# COOKIE_CHECK_SALT=$COOKIE_CHECK_SALT PASSWORD_SALT=$PASSWORD_SALT ADMIN_PASSWORD=$ADMIN_PASSWORD APP_ROOT=$APP_ROOT go run . 2>&1 # works

if [ $? -ne 0 ]; then
  echo "ConcertGetApp likely panicked and exited with a non-zero exit code"
fi