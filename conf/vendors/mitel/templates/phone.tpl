# Mitel Phone-Specific Configuration ({{ account.mac_address|lower }}.cfg)
# Phone: {{ phone.type }} (Model: {{ phone.model }})
# User: {{ account.phone_number }}

# Account / Line Settings
{%- for line in account.lines %}
{%- if line.type == "line" %}
# ------------- line {{line.number}} --------------
sip line{{line.number}} user name: {{ line.user_name|default:line.number }}
sip line{{line.number}} auth name: {{ line.auth_name|default:line.number }}
sip line{{line.number}} password: {{ line.password|default:domain.sip_password }}
sip line{{line.number}} display name: {{ line.display_name|default:line.screen_name|default:line.number }}
sip line{{line.number}} screen name: {{ line.screen_name|default:line.display_name|default:line.number }}
sip line{{line.number}} proxy: {{ line.registrar_ip|default:domain.sip_server_ip }}
sip line{{line.number}} proxy port: {{ line.registrar_port|default:domain.sip_server_port|default:5060 }}
sip line{{line.number}} registrar: {{ line.registrar_ip|default:domain.sip_server_ip }}
sip line{{line.number}} registrar port: {{ line.registrar_port|default:domain.sip_server_port|default:5060 }}
{%- endif %}
{%- endfor %}

# Individual settings generated from lines and features
{%- for cfg in keys_config %}
{{ cfg }}
{%- endfor %}

# Custom Overrides and Additional Settings
{%- if phone.settings %}
# Device-specific settings
{%- for key, val in phone.settings %}
{{ key }}: {{ val }}
{%- endfor %}
{%- endif %}