{% for domain in all_domains %}
  <!-- Domain: {{ domain.name }} -->
  {% for phone in domain.phones %}
    <!-- Phone: {{ phone.MacAddress }} -->
    <Description>{{ phone.Description }}</Description>
    <Number>{{ phone.PhoneNumber }}</Number>
    {% for line in phone.Lines %}
       {% set info = line.GetAdditionalInfoMap() %}
       <Entry>
         <Name>{{ info.display_name }}</Name>
         <UserName>{{ info.user_name }}</UserName>
         <Registrar>{{ info.registrar1_ip }}</Registrar>
       </Entry>
    {% endfor %}
  {% endfor %}
{% endfor %}