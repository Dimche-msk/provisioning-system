{# John,Smith,Acme Ltd.,Director of Marketing,123 Acme Rd.,Toronto,Ontario,L4K 4N9,Canada,,,,,,jsmith@acme.com,,,2,1,1,9054804321,3,1,9054801234 #}
{%- for domain in all_domains -%}
{%- for phone in domain.phones -%}
{%- if phone.Description and phone.PhoneNumber-%}
{%- set parts = phone.Description | split: " " -%}
{{ parts.0 }},{{ parts.1 | default: "" }},{{ domain.variables.company | default: "" }},{{ domain.variables.job_title | default: "" }},{{ domain.variables.work_street | default: "" }},{{ domain.variables.work_city | default: "" }},{{ domain.variables.work_state | default: "" }},{{ domain.variables.work_zip | default: "" }},{{ domain.variables.work_country | default: "" }},,,,,,{{domain.variables.email | default: "" }},,,1,1,1,{{ phone.PhoneNumber }},,,,
{% endif -%}
{%- endfor -%}
{%- endfor -%}
