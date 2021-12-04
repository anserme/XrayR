#!/usr/bin/env bash

sed -i "s|\"access\": \"/var/log/v2ray/access.log\",||g" "/etc/XrayR/config.yml"
sed -i "s|\"error\": \"/var/log/v2ray/error.log\",||g" "/etc/XrayR/config.yml"

getDomain(){
    domain=$(curl -s "${sspanel_url}/mod_mu/nodes/0/info?key=${key}&muKey=${key}" | grep -Eo -m 1 "[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+\.?" | head -1)
    echo $domain
}
echo "-----获取域名-----"
getDomain

if [ ! -z "${panel_type}" ]
    then
          sed -i "s|\"SSpanel\"|\"${panel_type}\"|"  "/etc/XrayR/config.yml"
fi

if [ ! -z "${sspanel_url}" ]
    then
         sed -i "s|\"http://127.0.0.1:667\"|\"${sspanel_url}\"|g" "/etc/XrayR/config.yml"
fi

if [ ! -z "${key}" ]
    then
         sed -i "s/\"55fUxDGFzH3n\"/\"${key}\"/g" "/etc/XrayR/config.yml"
fi

if [ ! -z "${node_id}" ]
    then
         sed -i "s/NodeID: 0/nodeid: ${node_id}/g" "/etc/XrayR/config.yml"
fi

if [ ! -z "${node_type}" ]
    then
         sed -i "s/NodeType: V2ray/nodeid: ${node_type}/g" "/etc/XrayR/config.yml"
fi

if [ ! -z "${timeout}" ]
    then
        sed -i "s/Timeout: 30/Timeout: ${timeout}/g" "/etc/XrayR/config.yml"
fi

if [ ! -z "${checkrate}" ]
    then
        sed -i "s/UpdatePeriodic: 60/UpdatePeriodic: ${checkrate}/g" "/etc/XrayR/config.yml"
fi

if [ ! -z "${enable_vless}" ]
    then
        sed -i "s/EnableVless: false/EnableVless: ${enable_vless}/g" "/etc/XrayR/config.yml"
fi

if [ ! -z "${enable_xtls}" ]
    then
        sed -i "s/EnableXTLS: false/EnableVless: ${enable_xtls}/g" "/etc/XrayR/config.yml"
fi

if [ ! -z "${cert_mode}" ]
    then
      sed -i "s|CertMode: dns|CertMode: ${cert_mode}|g" "/etc/XrayR/config.yml"
fi

if [ ! -z "${dns_mode}" ]
    then
      sed -i "s|Provider: cloudflare|Provider: ${dns_mode}|g" "/etc/XrayR/config.yml"
fi

if [ ! -z "${domain}" ]
    then
      sed -i "s|CertDomain: \"node1.test.com\"|CertDomain: \"${domain}\"|g" "/etc/XrayR/config.yml"
fi

if [ ! -z "${CF_Key}" ]
    then
  sed -i "s|CLOUDFLARE_API_KEY: bbb|CLOUDFLARE_API_KEY: ${CF_Key}|g" "/etc/XrayR/config.yml"
fi

if [ ! -z "${CF_Email}" ]
    then
  sed -i "s|CLOUDFLARE_EMAIL: aaa|CLOUDFLARE_EMAIL: ${CF_Email}|g" "/etc/XrayR/config.yml"

fi

if [ ! -z "${ALi_Key}" ]
    then
  sed -i "s|\"sdfsdfsdfljlbjkljlkjsdfoiwje\"|\"${ALi_Key}\"|g" "/etc/XrayR/config.yml"
fi

if [ ! -z "${Ali_Secret}" ]
    then
  sed -i "s|\"jlsdflanljkljlfdsaklkjflsa\"|\"${Ali_Secret}\"|g" "/etc/XrayR/config.yml"
fi

cat /etc/XrayR/config.yml
xrayr --config /etc/XrayR/config.yml
