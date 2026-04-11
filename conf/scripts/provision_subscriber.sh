#!/bin/bash

# Настройка переменных окружения (если требуется для MX-One)
export PATH=$PATH:/opt/eri_sn/bin/

usage() {
    echo "Использование: $0 -n <номер> [-f <полное_имя>] [-c <csp>] [-p <пароль>]"
    echo "  -n: Номер абонента (обязательно)"
    echo "  -f: Имя (одно слово или Имя Фамилия через пробел)"
    echo "  -c: CSP (по умолчанию 0)"
    echo "  -p: Пароль (auth_code)"
    exit 1
}

# Инициализация переменных по умолчанию
NUMBER=""
FULL_NAME=""
CSP=0
PASSWORD=""

# Парсинг именованных параметров
# getopts считывает флаги. Если флаг требует аргумента, за ним идет двоеточие.
while getopts "n:f:c:p:" opt; do
    case "$opt" in
        n) NUMBER="$OPTARG" ;;
        f) FULL_NAME="$OPTARG" ;;
        c) CSP="$OPTARG" ;;
        p) PASSWORD="$OPTARG" ;;
        *) usage ;;
    esac
done

# Проверка обязательного параметра номера
if [ -z "$NUMBER" ]; then
    echo "[ERROR] Номер абонента (-n) обязателен."
    usage
fi

# Обработка имени: делим по первому пробелу
NAME1=""
NAME2=""

if [[ "$FULL_NAME" == *" "* ]]; then
    # Если есть пробел: берем все до первого пробела в NAME1, все после — в NAME2
    NAME1="${FULL_NAME%% *}"
    NAME2="${FULL_NAME#* }"
else
    # Если пробела нет — всё имя идет в NAME1
    NAME1="$FULL_NAME"
    NAME2=""
fi

echo "--- Обработка абонента $NUMBER (CSP: $CSP) ---"
[ -n "$FULL_NAME" ] && echo "Данные имени -> Имя1: [$NAME1], Имя2: [$NAME2]"

# 1. Проверка и создание Extension
EXT_PRINT=$(/opt/eri_sn/bin/extension -p -d "$NUMBER" 2>/dev/null)
if echo "$EXT_PRINT" | grep -q "^$NUMBER "; then
    echo "[OK] Extension $NUMBER уже существует."
    
    # Сверяем CSP (4-я колонка в выводе extension -p)
    CURRENT_CSP=$(echo "$EXT_PRINT" | grep "^$NUMBER " | awk '{print $4}')
    if [ "$CURRENT_CSP" != "$CSP" ]; then
        echo "[..] Текущий CSP ($CURRENT_CSP) отличается от целевого ($CSP). Обновление..."
        /opt/eri_sn/bin/extension -e -d "$NUMBER" --csp "$CSP"
        echo "[OK] CSP обновлен до $CSP."
    fi
else
    echo "[..] Extension $NUMBER не найден. Создание с CSP $CSP..."
    /opt/eri_sn/bin/extension -i -d "$NUMBER" -l 1 --csp "$CSP" --max-terminals 1 --video yes --third-party-client yes --security-exception no
    
    if /opt/eri_sn/bin/extension -p -d "$NUMBER" 2>/dev/null | grep -q "^$NUMBER "; then
        echo "[OK] Extension $NUMBER успешно создан."
    else
        echo "[ERROR] Не удалось создать Extension $NUMBER."
        exit 1
    fi
fi

# 2. Проверка и создание IP Extension
IP_PRINT=$(/opt/eri_sn/bin/ip_extension -p -d "$NUMBER" 2>/dev/null)
if echo "$IP_PRINT" | grep -q "^$NUMBER "; then
    echo "[OK] IP Extension $NUMBER уже существует."
else
    echo "[..] IP Extension $NUMBER не найден. Создание..."
    /opt/eri_sn/bin/ip_extension -i -d "$NUMBER"
    if /opt/eri_sn/bin/ip_extension -p -d "$NUMBER" 2>/dev/null | grep -q "^$NUMBER "; then
        echo "[OK] IP Extension $NUMBER успешно создан."
    else
        echo "[ERROR] Не удалось создать IP Extension $NUMBER."
    fi
fi

# 3. Установка пароля (auth_code)
if [ -n "$PASSWORD" ]; then
    AUTH_PRINT=$(auth_code -p -d "$NUMBER" 2>/dev/null)
    if echo "$AUTH_PRINT" | grep -q " $NUMBER "; then
        # Пароль в auth_code -p находится в 3-й колонке
        CURRENT_AUTH=$(echo "$AUTH_PRINT" | grep " $NUMBER " | awk '{print $3}')
        if [ "$CURRENT_AUTH" == "$PASSWORD" ]; then
            echo "[OK] Пароль для $NUMBER уже установлен корректно."
        else
            echo "[..] Изменение пароля для $NUMBER..."
            /opt/eri_sn/bin/auth_code -e -d "$NUMBER"
            /opt/eri_sn/bin/auth_code -i -d "$NUMBER" --auth-code "$PASSWORD" --csp "$CSP" --cil "$NUMBER" --restricted
            echo "[OK] Пароль успешно изменен."
        fi
    else
        echo "[..] Установка пароля для $NUMBER..."
        /opt/eri_sn/bin/auth_code -i -d "$NUMBER" --auth-code "$PASSWORD" --csp "$CSP" --cil "$NUMBER" --restricted
        echo "[OK] Пароль успешно установлен."
    fi
fi

# 4. Проверка и создание/обновление Имени
if [ -n "$NAME1" ]; then
    NAME_PRINT=$(name -p --dir "$NUMBER" 2>/dev/null)
    CURRENT_NAME1=$(echo "$NAME_PRINT" | grep "^$NUMBER " | awk '{print $5}' | tr -d '"')
    CURRENT_NAME2=$(echo "$NAME_PRINT" | grep "^$NUMBER " | awk '{print $6}' | tr -d '"')

    if [ "$CURRENT_NAME1" == "$NAME1" ] && [ "$CURRENT_NAME2" == "$NAME2" ]; then
        echo "[OK] Имя для $NUMBER уже корректно."
    else
        echo "[..] Обновление имени: '$CURRENT_NAME1 $CURRENT_NAME2' -> '$NAME1 $NAME2'..."
        if [ -n "$NAME2" ]; then
            name -i -d "$NUMBER" --name1 "$NAME1" --name2 "$NAME2" --number-type dir
        else
            name -i -d "$NUMBER" --name1 "$NAME1" --number-type dir
        fi
        echo "[OK] Имя успешно обновлено."
    fi
fi

echo "--- Завершено для $NUMBER ---"
