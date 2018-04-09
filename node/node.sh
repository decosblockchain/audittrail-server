#!/bin/bash
if [ ! -f /data/config/node-initialized ]; then 
  echo "Generating node key..."
  bootnode -genkey  /data/config/node.key

  echo "Writing node pubkey..."
  bootnode -nodekey /data/config/node.key -writeaddress > /data/config/node.pub

  echo "Initializing blockchain..."
  geth --datadir "/data/blockchain" init /data/config/genesis.json

  touch /data/config/node-initialized
fi
geth --networkid "192001" --rpcapi "db,eth,net,web3" --rpcaddr $MINERIP --rpc --rpcport "8000" --port "33003" --datadir "/data/blockchain" --nodekey /data/config/node.key --nodiscover --verbosity 9  --exec "admin.addPeer('enode://301d36f1cee786f65cb527ba4bb7f2f330ae60b9a0cd8dbc1a547bbb05edf1bf3791d1ab6d70efb73980f83a9b9ccd551a726597b557be5316750558c4bbb64c@51.145.155.48:33003')"