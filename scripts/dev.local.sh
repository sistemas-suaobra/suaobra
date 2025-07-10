set -e

echo 'building binary'
rm -f suaobra-app && go build 

overmind start -r all