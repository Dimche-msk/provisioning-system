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
printf -v FORMATTED_PACKET "%b" "NOTIFY sip:$TARGET_IP SIP/2.0\r\n\
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

echo "Sending NOTIFY to $TARGET_IP:$TARGET_PORT..."

# Пытаемся отправить через /dev/udp (встроенная функция bash), если доступно
if bash -c "true > /dev/udp/$TARGET_IP/$TARGET_PORT" 2>/dev/null; then
    printf "%s" "$FORMATTED_PACKET" > /dev/udp/"$TARGET_IP"/"$TARGET_PORT"
    echo "Packet sent successfully via /dev/udp."
elif command -v nc >/dev/null 2>&1; then
    # Определяем версию nc и доступные флаги для кроссплатформенности
    NC_HELP=$(nc -h 2>&1 || nc --help 2>&1)
    if echo "$NC_HELP" | grep -q "\-q"; then
        # OpenBSD netcat
        printf "%s" "$FORMATTED_PACKET" | nc -u -q 0 "$TARGET_IP" "$TARGET_PORT"
    elif echo "$NC_HELP" | grep -q "\-c"; then
        # Ncat или традиционный netcat
        printf "%s" "$FORMATTED_PACKET" | nc -u -c "$TARGET_IP" "$TARGET_PORT"
    else
        # Fallback (используем таймаут 1 секунда)
        printf "%s" "$FORMATTED_PACKET" | nc -u -w 1 "$TARGET_IP" "$TARGET_PORT"
    fi
    echo "Packet sent successfully via nc."
else
    echo "Failed to send packet: neither bash /dev/udp nor nc found."
    exit 1
fi
