<flat-profile>
<Enable_Web_Server>           Yes       </Enable_Web_Server>
<Web_Server_Port>             80        </Web_Server_Port>
<Enable_Web_Admin_Access>     Yes       </Enable_Web_Admin_Access>
{%- if admin_password %}
<Admin_Passwd>{{admin_password}}</Admin_Passwd>
{%- endif %}
{%- if user_password %}
<User_Password>{{user_password}}</User_Password>
{%- endif %}

<!------ Provisioning params ------>
<Provision_Enable>Yes</Provision_Enable>
<Resync_On_Reset>yes</Resync_On_Reset>
<Resync_Periodic>3600</Resync_Periodic>
<Resync_Random_Delay>10</Resync_Random_Delay>
<Resync_From_SIP>yes</Resync_From_SIP>
<Profile_Rule>http://{{http_prov_server}}:{{http_prov_server_port}}/spa$MA.xml</Profile_Rule>

</flat-profile>
