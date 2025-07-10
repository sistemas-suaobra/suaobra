set -e 

# Build Binary & Data Assets
echo 'Building binary & Data Assets'
go build # TODO: use docker to build to maintain compatible GLIBC
chmod +x suaobra-app

mkdir -p data/core
mc find R2/sua-obra/staging --name "*.csv.gz" --exec "mc cp {} ./data/core/"
bash scripts/build.asset.sh

###################################################
# Push Binary & Data Assets
echo 'Push Binary & Data Assets'
mc cp ./suaobra-app R2/sua-obra/staging/
mc cp ./core.db.gz R2/sua-obra/staging/

###################################################
# Deploy to staging
echo 'Deploy to staging'

scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null .env $DEPLOY_SERVER:/__/git/suaobra-app/docker/staging/.env

ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $DEPLOY_SERVER "
export MC_HOST_R2="$MC_HOST_R2"

cd /__/storage/sua-obra-staging

# Get Data Assets
rm -f core.db*
mc cp R2/sua-obra/staging/core.db.gz .
gzip -d core.db.gz

# Get Binary
rm -f suaobra-app
mc cp R2/sua-obra/staging/suaobra-app .
chmod +x ./suaobra-app

cd /__/git/suaobra-app/
git pull
cd docker/staging

# to not interfere with load
docker restart dbnet

# Swap files
rm -f suaobra-app && mv /__/storage/sua-obra-staging/suaobra-app .
# rm -f core.db && mv /__/storage/sua-obra-staging/core.db .

docker compose up -d --build

# clean up
rm -f ./suaobra-app
docker system prune -f

# restart the app for refresh?
docker restart sua-obra-staging
"

# clean up
rm -f .env