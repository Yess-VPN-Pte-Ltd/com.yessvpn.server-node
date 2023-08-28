apt-get install -y curl
bash <(curl -L https://raw.githubusercontent.com/v2fly/fhs-install-v2ray/master/install-release.sh)
wget https://muxigame.github.io/deploy_shadowsocks/config.json  -O -> /usr/local/etc/v2ray/config.json
ufw enable
ufw allow 10086
ufw reload
service v2ray restart