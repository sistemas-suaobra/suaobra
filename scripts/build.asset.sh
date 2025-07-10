set -e

wget https://github.com/slingdata-io/sling-cli/releases/latest/download/sling_linux_amd64.tar.gz && tar -xf sling_linux_amd64.tar.gz && chmod 755 ./sling

mkdir -p ./data/core
./sling run -r store/sling/build.sqlite.core.yaml
mv ./data/core/core.db .

# optimize
sqlite3 ./core.db vacuum

# compress
gzip -f ./core.db