<flat-profile>
{%- if vlan_voip %}
<Enable_VLAN>Yes</Enable_VLAN>
<VLAN_ID>{{ vlan_voip }}</VLAN_ID>
<PC_Port_VLAN_ID>{{vlan_pc}}</PC_Port_VLAN_ID>
<Enable_PC_Port_VLAN_Tagging>No</Enable_PC_Port_VLAN_Tagging>
{%- endif %}

{%- for line in account.lines %}
{%- if line.type == "line" %}
<!-- line {{line.number}} -->
<User_ID_{{line.number}}_>{{ line.auth_name|default:line.phone_number }}</User_ID_{{line.number}}_>
<Password_{{line.number}}_>{{ line.password|default:domain.sip_password }}</Password_{{line.number}}_>
<Register_Expires_{{line.number}}_> 120 </Register_Expires_{{line.number}}_>
<Line_Enable_{{line.number}}_> Yes </Line_Enable_{{line.number}}_>
<SIP_Port_{{line.number}}_>{{ line.registrar1_port|default:domain.sip_server_port|default:5060 }}</SIP_Port_{{line.number}}_>
<Default_Ring_{{line.number}}_>7</Default_Ring_{{line.number}}_>
<Proxy_{{line.number}}_>{{ line.registrar1_ip|default:domain.sip_server_ip }}</Proxy_{{line.number}}_>
<Use_Outbound_Proxy_{{line.number}}_> No </Use_Outbound_Proxy_{{line.number}}_>
<Register_{{line.number}}_> Yes </Register_{{line.number}}_>
<Extension_{{line.number}}_>{{line.number}}</Extension_{{line.number}}_>
<Short_Name_{{line.number}}_>{{ line.display_name|default:line.screen_name|default:line.phone_number }}</Short_Name_{{line.number}}_>
<Dial_Plan_{{line.number}}_> (0[12345789]S0|[1-7]xxS0|8[2-9]xxxxxxxxxS0|810xx.|81[123456789]S0|*xx.|xx.) </Dial_Plan_{{line.number}}_>
<Preferred_Codec_{{line.number}}_>G711a</Preferred_Codec_{{line.number}}_>
<Use_Pref_Codec_Only_{{line.number}}_>Yes   </Use_Pref_Codec_Only_{{line.number}}_>
{%- endif %}
{%- endfor %}



<Enable_CDP>No</Enable_CDP>


<!-- Phone Input Gains -->

<Handset_Input_Gain> 6 </Handset_Input_Gain>
<Headset_Input_Gain> 6 </Headset_Input_Gain>

<Dial_Tone>420@-19,420@-19;10(*/0/1+2)</Dial_Tone>
<Outside_Dial_Tone>420@-16;10(*/0/1)</Outside_Dial_Tone>
<Busy_Tone>420@-19,420@-19;10(.5/.5/1+2)</Busy_Tone>
<Ring_Back_Tone>420@-19,420@-19;*(2/4/1+2)</Ring_Back_Tone>
<DND_Serv>No</DND_Serv>
<!-- User Supplementary Services -->

<Time_Format> 12hr </Time_Format>
<Date_Format> day/month </Date_Format>


<!-- Time Zone Selection -->
<Daylight_Saving_Time_Enable>No</Daylight_Saving_Time_Enable>
<Time_Zone>GMT+03:00 </Time_Zone>

<!-- System Configuration -->

<Enable_Web_Server>           Yes       </Enable_Web_Server>
<Web_Server_Port>             80        </Web_Server_Port>
<Enable_Web_Admin_Access>     Yes       </Enable_Web_Admin_Access>
<Admin_Passwd>                			</Admin_Passwd>
<User_Password>                         </User_Password>
<Resync_Periodic>       3520      </Resync_Periodic>
<Resync_Random_Delay>       600         </Resync_Random_Delay>

<!-- Optional Network Configuration -->

<Primary_DNS>      {{domain.dns_server|default:"8.8.8.8"}}  </Primary_DNS>
<Secondary_DNS>    {{domain.dns_server2|default:"8.8.4.4"}}  </Secondary_DNS>
<DNS_Server_Order> Manual         </DNS_Server_Order>
<DNS_Query_Mode>   Parallel       </DNS_Query_Mode>

<Primary_NTP_Server>{{domain.ntp_server|default:ru.pool.ntp.org}}</Primary_NTP_Server>
{%- if domain.ntp_server2 %}
<Secondary_NTP_Server>{{domain.ntp_server2|default:ru.pool.ntp.org}}</Secondary_NTP_Server>
{%- endif %}
<Idle_Key_List>em_login|1;acd_login|1;acd_logout|1;avail|3;unavail|3;redial|5;dir|6;cfwd|7;|8;lcr|9;pickup|10;gpickup|11;unpark|12;em_logout</Idle_Key_List>
<Group_Paging_Script group="Phone/Multiple_Paging_Group_Parameters"></Group_Paging_Script>
</flat-profile>
