#!version:1.0.0.1
#File header "#!version:1.0.0.1" cannot be edited or deleted.##
{# ---------------------------------Generate Line Config ----------------------------------#}
{%- for line in account.lines %}
{%- if line.type == "line" %}
# ------------- line {{line.number}} --------------
account.{{line.number}}.enable = 1
account.{{line.number}}.label = {{ line.screen_name|default:line.display_name|default:line.number }}
account.{{line.number}}.display_name = {{ line.display_name|default:line.screen_name|default:line.number }}
account.{{line.number}}.auth_name = {{ line.auth_name|default:line.number }}
account.{{line.number}}.user_name = {{ line.user_name|default:line.number }}
account.{{line.number}}.password = {{ line.password|default:domain.sip_password }}
{%- if line.custom_sip_server %}
account.{{line.number}}.sip_server.1.address = {{line.custom_sip_server.registrar_ip}}
account.{{line.number}}.sip_server.1.port = {{line.custom_sip_server.registrar_port}}
account.{{line.number}}.outbound_proxy.1.address = {{line.custom_sip_server.proxy_ip}}
account.{{line.number}}.outbound_proxy.1.port = {{line.custom_sip_server.proxy_port}}
{%- endif %}
{%- endif %}
{%- endfor %}
{# ---------------------------------Generate Keys Config ----------------------------------#}

#------------------  KEYS ----------------
{%- for cfg in keys_config %}
{{ cfg }}
{%- endfor %}
