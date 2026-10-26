-- Docker Build
docker build -f NextDuel/Dockerfile -t nextduel:latest .

-- Docker Export
docker save -o nextduel.tar nextduel:latest

-- Docker Load
docker load -i nextduel.tar