set -e 

###################################################
# Push Binary & Data Assets
echo 'Push Binary & Data Assets'
mc cp R2/sua-obra/staging/suaobra-app R2/sua-obra/production/
# mc cp R2/sua-obra/staging/core.db.gz R2/sua-obra/production/ # getting All non-trailing parts must have the same length.

###################################################
# Deploy to production
echo 'Deploy to production'

scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null .env $DEPLOY_SERVER:/__/git/suaobra-app/docker/production/.env

ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $DEPLOY_SERVER "
export MC_HOST_R2="$MC_HOST_R2"

cd /__/storage/sua-obra-production

# Get Data Assets
rm -f core.db*
mc cp R2/sua-obra/staging/core.db.gz .
gzip -d core.db.gz

# Get Binary
rm -f suaobra-app
mc cp R2/sua-obra/production/suaobra-app .
chmod +x ./suaobra-app

cd /__/git/suaobra-app/
git pull
cd docker/production

# to not interfere with load
docker restart dbnet

# Swap files
rm -f suaobra-app && mv /__/storage/sua-obra-production/suaobra-app .
rm -f core.db && mv /__/storage/sua-obra-production/core.db .

docker compose up -d --build

# clean up
rm -f ./suaobra-app
docker system prune -f

# restart the app for refresh?
docker restart sua-obra-production
"

# clean up
rm -f .env