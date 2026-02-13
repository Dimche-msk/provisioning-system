# SIP Line 1
sip line1 user name:{{account.settings.account1_name}} #Enter user extension number
#If password is defined via auth_code_init
sip line1 password:12345
sip line1 auth name:{{account.auth_name}}

sip proxy ip:{{domain.sip_proxy_address}}
sip proxy port:{{domain.sip_proxy_port}}
sip registrar ip:{{domain.sip_registrar_address}}
sip registrar port:{{domain.sip_registrar_port}}


#directory disabled: 0
directory 1: Corporate.csv
#directory 2: Private319.csv
directory 1 name:Корпоративный
#directory 2 name:Личный


{# ---------------------------------Generate Keys Config ----------------------------------#}

#------------------  KEYS ----------------
{%- for cfg in keys_config %}
{{ cfg }}
{%- endfor %}