# Mitel Global Configuration (aastra.cfg)
# Generated for Domain: {{ variables.domain_name }}

# Network Settings
dhcp: 1
download protocol: HTTP
http server: {{ variables.http_prov_server }}
http path: /
http port: 80

# Time and Date
time server disabled: 0
time server1: {{ variables.ntp_server }}
time zone name: {{ variables.time_zone|default:"Europe/Moscow" }}
time format: 1 # 24h
date format: 2 # YYYY-MM-DD

# Localization
tone set: Europe
input language: Russian

# Global SIP Settings
sip transport protocol: 1 # UDP
sip customized codec: payload=9;ptime=20;silsupp=off,payload=8;ptime=20;silsupp=off,payload=0;ptime=20;silsupp=off
sip silence suppression: 0
sip out-of-band dtmf: 1
sip dtmf method: 1 # SIP INFO

# Security
web interface enabled: 1
admin password: {{ variables.admin_password|default:"22222" }}

# Feature codes
directed call pickup: 1
directed call pickup prefix: Pickup
conference disabled: 1

# Default registration settings
auto resync mode: 3
auto resync time: 03:00

# Expansion Module Settings (if any)
# ...
