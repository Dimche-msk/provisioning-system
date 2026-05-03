# Mitel Global Configuration (aastra.cfg)
# Generated for Domain: {{ variables.domain_name }}

# Network Settings
dhcp: 1
download protocol: HTTP
http server: {{ variables.http_prov_server }}
http path: /
http port: 8080

# Time and Date
time server disabled: 0
time server1: {{ variables.ntp_server }}
time zone name: {{ variables.time_zone|default:"RU-Moscow" }}
time format: 1 # 24h
date format: 2 # YYYY-MM-DD

# Localization
tone set: Europe
input language: Russian

# Global SIP Configuration
sip proxy ip: {{variables.sip_server1_ip}}
sip proxy port:{{variables.sip_server1_port|default:5060}}
sip registrar ip:0.0.0.0
sip registrar port:5060
sip line1 user name:"Not Configured"
sip backup proxy port:5060
sip backup registrar port:5060
dynamic sip:1
sip subscription timeout retry timer:300
sip subscription failed retry timer:300
sip registration retry timer:300
sip gruu:0
sip session timer:1800
sip rtcp summary reports:0
sip rtcp summary report collector port:0
sip srtp mode:0
sips persistent tls:0
sip persistent tls keep alive:0
sips symmetric tls signaling:1
sip send sips over tls:1
sip outbound support:0
sip outbound proxy:0.0.0.0
sip outbound proxy port:5061
sip whitelist:0

sip dial plan:"x+^|xx+*"
sip dial plan terminator:1
sip digit timeout:3

sip aastra id:1
sip send line:1
sip xml notify event:1
sip pai:1
sip vmail:"*32#"
sip explicit mwi subscription:0
sip services transport protocol:-1

sip line1 dtmf method:1
sip line2 dtmf method:1
sip allow auto answer:1
sip ignore status code:603
sip intercom allow barge in:0
sip intercom mute mic:1
sip intercom warning tone:1

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
#---------------------------------------------------------
# softkey selection list page A-269
#softkey selection list: blf, speeddial, line, xml
# values 
# none
# filter
# line
# callers
# speeddial
# redial
# dnd
# conf
# blf
# xfer
# list
# icom
# acd
# services
# xml
# phonelock
# flash
# paging
# spre
# save
# park
# delete
# pickup
# hotdesklogin
# lcr
# discreetringing
# callforward
# callhistory
# blfxfer
# mystatus
# speeddialxfer
# contacts
# speeddialconf
# favorite
# speeddialmwi
# empty
# directory
#