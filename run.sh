#!/bin/bash

l1_rpc_from_input=$1
snap_flag_from_input=$2
need_snap=false

function fix_l1_rpc() {
	env_l1rpc="XGON_NODE_ETHERMAN_URL =  \"$l1_rpc_from_input\""
	cp example.env .env
	echo $env_l1rpc >>.env
}

function check_use_snap() {
	if [ "$snap_flag_from_input" == "snap" ]; then
		need_snap=true
	fi
}

function linux_and_mac() {
	wget https://static.okex.org/cdn/chain/xgon/snapshot/testnet.zip && unzip testnet.zip && cd testnet
	fix_l1_rpc

	if $need_snap == true; then
		wget https://static.okex.org/cdn/chain/xgon/snapshot/testnet-latest
		latest_snap=$(cat testnet-latest)
		wget https://static.okex.org/cdn/chain/xgon/snapshot/"$latest_snap"
		tar -zxvf $latest_snap
		docker-compose --env-file .env -f ./docker-compose.yml up -d xgon-state-db
		sleep 5
		docker run --network=xgon -v "$(pwd)":/data okexchain/xgon-node:origin_release_v0.1.0_20231107071509 /app/xgon-node restore --cfg /data/config/node.config.toml -is /data/xgon-testnet-snp/state_db.sql.tar.gz -ih /data/xgon-testnet-snp/prover_db.sql.tar.gz
		sleep 5
	else
		echo "not need snap"
	fi

	docker-compose --env-file .env -f ./docker-compose.yml up -d
}

function windows() {
	curl -LO https://static.okex.org/cdn/chain/xgon/snapshot/testnet.zip
	tar -xf testnet.zip
	cd testnet
	fix_l1_rpc

	if $need_snap == true; then
		curl -LO https://static.okex.org/cdn/chain/xgon/snapshot/testnet-latest
		latest_snap=$(cat testnet-latest)
		curl -LO https://static.okex.org/cdn/chain/xgon/snapshot/"$latest_snap"
		tar -zxvf $latest_snap
		docker-compose --env-file .env -f ./docker-compose.yml up -d xgon-state-db
		timeout /t 5
		set CURRENT_DIR=%cd%
		docker run --network=xgon -v "%CURRENT_DIR%":/data okexchain/xgon-node:origin_release_v0.1.0_20231107071509 /app/xgon-node restore --cfg /data/config/node.config.toml -is /data/xgon-testnet-snp/state_db.sql.tar.gz -ih /data/xgon-testnet-snp/prover_db.sql.tar.gz
		timeout /t 5
	else
		echo "not need snap"
	fi

	docker-compose --env-file .env -f ./docker-compose.yml up -d
}

check_use_snap

uNames=$(uname -s)
osName=${uNames:0:4}
if [ "$osName" == "Darw" ]; then # Darwin
	echo "Mac OS X"
	linux_and_mac
elif [ "$osName" == "Linu" ]; then # Linux
	echo "GNU/Linux"
	linux_and_mac
elif [ "$osName" == "MING" ]; then # MINGW, windows, git-bash
	echo "Windows, git-bash"
	# windows
else
	echo "unknown os"
fi

