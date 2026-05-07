#!/bin/bash

# Скрипт для получения IP-адреса расширения из Mitel MX-One
# Использование: ./get_mxone_ip.sh <номер_телефона>

if [ -z "$1" ]; then
    echo "Usage: $0 <extension_number>"
    exit 1
fi

EXTENSION="$1"

# Выполняем команду и извлекаем IP
# Мы ищем строку, которая начинается с номера расширения, и берем 3-й столбец
IP=$(ip_extension_info -d "$EXTENSION" -f reg 2>/dev/null | grep "^$EXTENSION" | awk '{print $3}')

if [ -n "$IP" ]; then
    # Выводим только чистый IP
    echo "$IP"
else
    # Если регистрация не найдена, выходим с ошибкой
    exit 1
fi
