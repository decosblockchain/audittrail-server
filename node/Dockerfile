FROM ubuntu:latest
RUN apt-get update && apt-get install -y software-properties-common \
      && add-apt-repository -y ppa:ethereum/ethereum \
      && apt-get update \
      && apt-get -y install ethereum

COPY genesis.json /data/config/genesis.json
COPY connect.js /data/config/connect.js
ADD node.sh /root/scripts/node.sh
RUN chmod +x /root/scripts/node.sh
ENTRYPOINT ["/root/scripts/node.sh"]
