account.1.label = {{account.phone_number}}
account.1.display_name = {{account.settings.account1_name}}
domain = {{variables.http_prov_server}}
sip_password = {% if account.lines.0.password -%}{{ account.lines.0.password }}{%- elif variables.PasswdPre and variables.PasswdPost and (account.lines.0.auth_name or account.lines.0.phone_number) -%}{{ variables.PasswdPre }}{{ account.lines.0.auth_name|default:account.lines.0.phone_number }}{{ variables.PasswdPost }}{%- else -%}{{ variables.sip_password }}{%- endif %}

