# Provisioning System

## 1. Назначение системы

**Provisioning System** — это централизованная система для автоматизированного управления конфигурациями IP-телефонов. Она решает задачу массовой настройки и сопровождения парка телефонных аппаратов различных производителей.

**Система полностью сгенерирована ИИ от Google "Gemini 3".** 
Эта разработка стала тестом модели, с которым модель прекрасно справилась. 
Модели была предложена архитектура, которая подходила под мои текущие задачи, далее модели давались указания реализовать тот или иной функционал или исправить ошибки.
За все время разработки, мне пришлось только пару раз "прикоснуться" к коду, все остальное сделал ИИ

Система предназначена для развертывания на сервере и управляется через web-интерфейс

Основные возможности системы:
*   **Мультивендорность**: Поддержка любых моделей телефонов (Cisco, Yealink, Fanvil и др.) через гибкую систему шаблонов (Jinja2-like синтаксис).
*   **Мультидоменность**: Возможность ведения независимых конфигураций для разных филиалов или клиентов (доменов).
*   **Веб-интерфейс**: Удобный UI для добавления телефонов, управления линиями, настройки кнопок (BLF, Speed Dial) и импорта данных из Excel.
*   **Генерация конфигураций**: Автоматическое создание конфигурационных файлов при изменении настроек.
*   **Телефонные справочники**: Автоматическая генерация и обновление корпоративных телефонных книг (XML Directory) для каждого домена.
*   **Автоматизация**: Поддержка скриптов деплоя (Deploy Commands) для автоматической доставки конфигураций на TFTP/HTTP сервера. Возможность выполнения команд, создающих или удаляющих телефоны на АТС при изменении данных в системе (Автоматизация АТС)
*   **Генерация паролей**: Опциональная автоматическая генерация безопасных SIP-паролей.
*   **Поддержка многоязычности**: Интерфейс системы работает на русском и английском языке. Есть возможность добавления других языков.  
*   **Авторизация по логин/пароль**: Текущий релиз поддерживает авторизованный доступ администратора  



## 2. Общее описание развертывания системы

Исходный код проекта доступен в репозитории: [https://github.com/Dimche-msk/provisioning-system.git](https://github.com/Dimche-msk/provisioning-system.git)

Система поставляется в виде скомпилированного исполняемого файла (single binary), который содержит в себе и бэкенд, и фронтенд.

### Установка из готовых релизов

1.  **Скачивание**:
    Перейдите в раздел [Releases](https://github.com/Dimche-msk/provisioning-system/releases) на GitHub и скачайте архив с исполняемым файлом для вашей ОС:
    *   `provisioning-system-linux-amd64` (Linux 64-bit)
    *   `provisioning-system-linux-386` (Linux 32-bit)
    *   `provisioning-system-windows-amd64.exe` (Windows 64-bit)

2.  **Структура директорий**:
    Создайте рабочую папку для приложения и разместите в ней:
    *   Исполняемый файл приложения.
    *   Папку `conf/`, содержащую:
        *   `provisioning-system.yaml` — основной файл конфигурации.
        *   `vendors/` — директория с шаблонами производителей.

    Пример структуры:
    ```text
    /opt/provisioning-system/
    ├── provisioning-system-linux-amd64
    └── conf/
        ├── provisioning-system.yaml
        └── vendors/
            ├── cisco/
            │   ├── models/
            │   ├── templates/
            │   ├── directory/
            │   └── vendor.yaml
            └── yealink/
                └── ...
    ```

3.  **Конфигурация**:
    Настройте `conf/provisioning-system.yaml`, указав параметры базы данных, порта веб-сервера и настройки доменов.

4.  **Запуск**:
    *   **Linux**:
        ```bash
        chmod +x provisioning-system-linux-amd64
        ./provisioning-system-linux-amd64
        ```
        *Рекомендуется запускать через systemd как сервис.*

    *   **Windows**:
        Запустите `provisioning-system-windows-amd64.exe` через командную строку или как службу.

5.  **Использование**:
    Откройте браузер и перейдите по адресу `http://<адрес-сервера>:8080` (порт по умолчанию). Логин и пароль задаются в конфигурационном файле.

### Сборка из исходного кода

Если вы хотите собрать проект самостоятельно:

**Требования**:
*   Go 1.21 или выше
*   Node.js 18 или выше (для сборки фронтенда)

**Шаги сборки**:

1.  Клонируйте репозиторий:
    ```bash
    git clone https://github.com/Dimche-msk/provisioning-system.git
    cd provisioning-system
    ```

2.  Соберите фронтенд:
    ```bash
    cd frontend
    npm install
    npm run build
    cd ..
    ```
    *Скомпилированные файлы фронтенда будут помещены в `backend/cmd/server/static`.*

3.  Соберите бэкенд:
    ```bash
    cd backend
    go build -o provisioning-system cmd/server/main.go
    ```

4.  Готовый бинарный файл `provisioning-system` (или `.exe`) появится в папке `backend`.

## 3. Конфигурация системы

Основной файл конфигурации `conf/provisioning-system.yaml` управляет параметрами сервера, базы данных, авторизации и настройками доменов.

### Основные секции

*   **server**: Настройки веб-сервера.
    *   `port`: Порт, на котором будет доступен веб-интерфейс (по умолчанию "8090").
    *   `serve_configs`: Если `true`, система сама отдает сгенерированные файлы конфигурации по HTTP.
    *   `log_device_access`: Уровень логирования доступа устройств (`none`, `access`, `error`, `full`).
    *   `log_file_path`: Путь к файлу логов доступа.

*   **auth**: Настройки безопасности.
    *   `admin_user`: Логин администратора.
    *   `admin_password`: Пароль администратора.
    *   `secret_key`: Секретный ключ для подписи сессий (JWT).

*   **database**: Настройки базы данных.
    *   `path`: Путь к файлу базы данных SQLite (например, `provisioning.db`).
    *   `backup_dir`: Папка для хранения автоматических бэкапов базы данных.

*   **domains**: Список доменов (филиалов/клиентов). Каждый домен имеет свои уникальные настройки и переменные.
    *   `name`: Уникальное имя домена (используется в URL и путях к файлам).
    *   `deploy_commands`: Список команд, выполняемых после генерации конфигурации (например, копирование файлов на TFTP сервер, перезагрузка АТС). Поддерживает шаблонизацию.
    *   `delete_commands`: Список команд, выполняемых при удалении телефона.
    *   `generate_random_password`: Если `True`, система автоматически генерирует пароль для новых телефонов, если он не задан.
    *   `variables`: Произвольные переменные (ключ-значение), которые доступны в шаблонах конфигурации (например, IP адрес SIP сервера, NTP сервер, VLAN и т.д.).

### Пример конфигурации

```yaml
server:
  port: "8090"
  serve_configs: true
  log_device_access: full
  log_file_path: ./access.log

auth:
  admin_user: "admin"
  admin_password: "password123"
  secret_key: "change-me-to-something-secure"

database:
  path: "provisioning.db"
  backup_dir: "backups"

domains:
  - name: "MainOffice"
    generate_random_password: True
    deploy_commands:
      - "scp {{.FilePath}} user@tftp-server:/tftpboot/"
      - "ssh user@asterisk 'asterisk -rx \"sip reload\"'"
    variables:
      ntp_server: "pool.ntp.org"
      sip_server_ip: "192.168.1.10"
      sip_server_port: "5060"
      timezone: "Europe/Moscow"

  - name: "BranchOffice"
    variables:
      ntp_server: "pool.ntp.org"
      sip_server_ip: "10.10.1.10"
      sip_server_port: "5060"
```

## 4. Настройка вендоров (Vendor Configuration)

Система позволяет легко добавлять поддержку новых производителей телефонов. Конфигурация каждого вендора находится в отдельной папке внутри `conf/vendors/` (например, `conf/vendors/cisco/`).

### Файл vendor.yaml

В корне папки вендора должен находиться файл `vendor.yaml`, описывающий основные параметры.

Пример (`conf/vendors/cisco/vendor.yaml`):

```yaml
id: cisco                     # Уникальный ID вендора
name: Cisco                   # Отображаемое имя
static_dir: static            # Папка со статическими файлами (прошивки, картинки), которые будут скопированы как есть
phone_config_file: "spa{{account.mac_address}}.xml" # Шаблон имени файла конфигурации
phone_config_template: templates/phone.tpl          # Путь к основному шаблону конфигурации
```

### Шаблоны (Templates)

Шаблоны используют синтаксис **Pongo2** (аналог Jinja2 для Go). Они позволяют динамически формировать конфигурационные файлы на основе данных о телефоне и домене.

#### Доступные переменные в шаблонах

В шаблонах доступны следующие объекты:

1.  **`account`** — Объект текущего телефона:
    *   `mac_address`: MAC-адрес телефона.
    *   `phone_number`: Основной номер телефона.
    *   `lines`: Список линий телефона.
        *   `number`: Номер линии (1, 2, ...).
        *   `auth_name`: Имя для авторизации (из Additional Info).
        *   `password`: Пароль (из Additional Info).
        *   `display_name`: Отображаемое имя.
        *   `registrar1_ip`, `registrar1_port`: Адрес SIP сервера.
    *   `max_lines`: Максимальное количество линий для данной модели.
    *   `lines_range`: Массив номеров линий (для итерации).
    *   `used_line_numbers`: Карта использованных номеров линий.

2.  **`domain`** — Объект текущего домена:
    *   `name`: Имя домена.
    *   Любые переменные, заданные в `provisioning-system.yaml` в секции `variables` (например, `domain.sip_server_ip`, `domain.ntp_server`).

3.  **`phones`** — Список всех телефонов в текущем домене (полезно для генерации справочников).

4.  **`all_domains`** — Список всех доменов и их телефонов (для глобальных справочников).

#### Пример шаблона (Cisco)

```xml
<flat-profile>
    <!-- Настройки VLAN из переменных домена -->
    {%- if domain.vlan_voip %}
    <Enable_VLAN>Yes</Enable_VLAN>
    <VLAN_ID>{{ domain.vlan_voip }}</VLAN_ID>
    {%- endif %}

    <!-- Итерация по линиям телефона -->
    {%- for line in account.lines %}
    {%- if line.type == "line" %}
    <!-- Линия {{line.number}} -->
    <User_ID_{{line.number}}_>{{ line.auth_name|default:line.phone_number }}</User_ID_{{line.number}}_>
    <Password_{{line.number}}_>{{ line.password|default:domain.sip_password }}</Password_{{line.number}}_>
    <Proxy_{{line.number}}_>{{ line.registrar1_ip|default:domain.sip_server_ip }}</Proxy_{{line.number}}_>
    {%- endif %}
    {%- endfor %}

    <!-- Глобальные настройки -->
    <Primary_NTP_Server>{{domain.ntp_server|default:"pool.ntp.org"}}</Primary_NTP_Server>
</flat-profile>
```

### Генерация справочников (Directory)

Для автоматической генерации телефонных книг создайте папку `directory` внутри папки вендора (например, `conf/vendors/cisco/directory/`). Поместите туда шаблон (например, `directory.xml.tpl`).

Этот шаблон будет автоматически обрабатываться при любом изменении телефонов, и результат будет сохранен в корень конфигурации домена.

Пример шаблона справочника:
```xml
<Directory>
{% for phone in phones %}
  <Entry>
    <Name>{{ phone.Description }}</Name>
    <Number>{{ phone.PhoneNumber }}</Number>
  </Entry>
{% endfor %}
</Directory>
```
