{% if phone.type != "gateway" %}
# Mitel Phone-Specific Configuration ({{ account.mac_address|lower }}.cfg)
# Phone: {{ phone.type }} (Model: {{ phone.model }})
# User: {{ account.phone_number }}

# Account / Line Settings
{%- for line in account.lines %}
{%- if line.type == "line" %}
# ------------- line {{line.number}} --------------
sip line{{line.number}} user name: {{ line.user_name|default:line.number }}
sip line{{line.number}} auth name: {{ line.auth_name|default:line.number }}
sip line{{line.number}} password: {% if line.password -%}{{ line.password }}{%- elif variables.PasswdPre and variables.PasswdPost and (line.auth_name or line.phone_number) -%}{{ variables.PasswdPre }}{{ auth_name|default:line.phone_number }}{{ variables.PasswdPost }}{%- else -%}{{ variables.sip_password }}{%- endif %}
sip line{{line.number}} display name: {{ line.display_name|default:line.screen_name|default:line.number }}
sip line{{line.number}} screen name: {{ line.screen_name|default:line.display_name|default:line.number }}
sip line{{line.number}} proxy: {{ line.registrar_ip|default:variables.sip_server_ip }}
sip line{{line.number}} proxy port: {{ line.registrar_port|default:variables.sip_server_port|default:5060 }}
sip line{{line.number}} registrar: {{ line.registrar_ip|default:variables.sip_server_ip }}
sip line{{line.number}} registrar port: {{ line.registrar_port|default:variables.sip_server_port|default:5060 }}
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
{% else %}
# This device ({{ account.mac_address }}) is a {{ phone.type }} and does not require an individual configuration file.
# It is used for system number tracking and directory generation only.
{% endif %}