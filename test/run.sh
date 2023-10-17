# global constant
ZKEVM_NODE_STATE_DB_HOST="127.0.0.1:5432"
ZKEVM_NODE_POOL_DB_HOST="127.0.0.1:5433"
ZKEVM_NODE_SEQUENCER_SENDER_ADDRESS=0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266
ZKEVM_NODE_AGGREGATOR_SENDER_ADDRESS=0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266

# run zkevm-approve
echo 'start zkevm-approve'
./zkevm-node approve --network custom --custom-network-file ./config/test.genesis.config.json --key-store-path ./sequencer.keystore --pw testonly --am 115792089237316195423570985008687907853269984665640564039457584007913129639935 -y --cfg ./config/local.test.node.config.toml

sleep 5

##run zkevm-sync
#echo 'start zkevm-sync'
#nohup ./zkevm-node run --network custom --custom-network-file ./config/test.genesis.config.json --cfg ./config/local.test.node.config.toml --components "synchronizer" > xgon-sync.log 2>&1 &
#
#sleep 2

#run zkevm-node
echo 'start zkevm-node'
nohup ./zkevm-node run --network custom --custom-network-file ./config/test.genesis.config.json --cfg ./config/local.test.node.config.toml --components "synchronizer,eth-tx-manager,sequencer,sequence-sender,l2gaspricer,aggregator,rpc" > xgon-node.log 2>&1 &
