#!/bin/bash

# Проверка аргументов
if [ "$#" -lt 2 ]; then
    echo "Usage: $0 <target_ip> <my_ip> [target_port]"
    exit 1
fi

TARGET_IP="$1"
MY_IP="$2"
TARGET_PORT="${3:-5060}" # Порт по умолчанию 5060

# Генерируем UUID-подобные строки с дефисами (как в Python)
gen_uuid() {
    LC_ALL=C tr -dc 'a-f0-9' < /dev/urandom | head -c 32 | sed -E 's/([0-9a-f]{8})([0-9a-f]{4})([0-9a-f]{4})([0-9a-f]{4})([0-9a-f]{12})/\1-\2-\3-\4-\5/'
}

CALL_ID="$(gen_uuid)@$MY_IP"
FROM_TAG="$(LC_ALL=C tr -dc 'a-f0-9' < /dev/urandom | head -c 8)"
BRANCH="z9hG4bK-$(gen_uuid)"

# Формируем пакет с явными \r\n
SIP_PACKET="NOTIFY sip:$TARGET_IP SIP/2.0\r\n\
Via: SIP/2.0/UDP $MY_IP:5060;branch=$BRANCH\r\n\
From: <sip:provisioning@$MY_IP>;tag=$FROM_TAG\r\n\
To: <sip:$TARGET_IP>\r\n\
Call-ID: $CALL_ID\r\n\
CSeq: 1 NOTIFY\r\n\
Contact: <sip:provisioning@$MY_IP>\r\n\
Event: check-sync;reboot=false\r\n\
Max-Forwards: 70\r\n\
Subscription-State: terminated\r\n\
Content-Length: 0\r\n\r\n"

echo "Sending NOTIFY to $TARGET_IP:$TARGET_PORT via nc..."

# Используем nc (netcat) для отправки ОДНИМ пакетом
# -c (или --close) закроет соединение после передачи данных
printf "$SIP_PACKET" | nc -u -c "$TARGET_IP" "$TARGET_PORT"

if [ $? -eq 0 ]; then
    echo "Packet sent successfully."
else
    echo "Failed to send packet."
fi
