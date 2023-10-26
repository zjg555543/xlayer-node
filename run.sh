wget https://github.com/okx/Xgon-node/raw/scf/snap/xgon_devnet.zip && unzip xgon_devnet.zip && cd xgon_devnet_config

wget https://github.com/okx/Xgon-node/raw/scf/snap/aprover_db_1697611693_v0.2.6-RC3-18-gd16d4a42_d16d4a42.sql.tar.gz
wget https://github.com/okx/Xgon-node/raw/scf/snap/astate_db_1697611693_v0.2.6-RC3-18-gd16d4a42_d16d4a42.sql.tar.gz


rm -rf ./xagon_devnet_test
docker-compose --env-file .env -f ./docker-compose.yml up -d zkevm-state-db
sleep 5

docker run --network=zkevm -v ./:/data  okexchain/xagon-node:origin_release_v0.1.0_20231013105844 /app/xagon-node restore --cfg /data/node.config.toml -is /data/astate_db_1697611693_v0.2.6-RC3-18-gd16d4a42_d16d4a42.sql.tar.gz -ih /data/aprover_db_1697611693_v0.2.6-RC3-18-gd16d4a42_d16d4a42.sql.tar.gz
sleep 2

docker-compose --env-file .env -f ./docker-compose.yml up -d
