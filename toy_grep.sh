set -e # Exit early if any commands fail
(
  cd "$(dirname "$0")" # Ensure compile steps are run within the repository directory
  echo "\e[32mBuilding\e[0m"
  go build -o /tmp/codecrafters-build-grep-go app/*.go
  echo "\e[32mBuild Complete\e[0m"
)

exec /tmp/codecrafters-build-grep-go "$@"
