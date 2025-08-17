set -e

echo "Passo 1: Executando o ETL em Go para processar e enriquecer os dados..."
go run cmd/etl/main.go

echo "Passo 2: Baixando e configurando o Sling para carregar os dados..."
curl -sL https://github.com/slingdata-io/sling-cli/releases/latest/download/sling_darwin_arm64.tar.gz -o sling_darwin_arm64.tar.gz && tar -xf sling_darwin_arm64.tar.gz && chmod 755 ./sling

mkdir -p ./data/core
# O arquivo sling já espera os CSVs em data/core/, então apenas executamos.
./sling run -r store/sling/build.sqlite.core.yaml --log-level debug
mv ./data/core/core.db .

echo "Passo 3: Otimizando o banco de dados..."
sqlite3 ./core.db "vacuum;"

echo "Passo 4: Comprimindo o banco de dados final..."
gzip -f ./core.db

echo "Build de assets concluído com sucesso!"
