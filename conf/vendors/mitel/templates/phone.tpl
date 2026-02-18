# Mitel Phone-Specific Configuration ({{ account.mac_address|lower }}.cfg)
# Phone: {{ phone.type }} (Model: {{ phone.model }})
# User: {{ account.phone_number }}

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