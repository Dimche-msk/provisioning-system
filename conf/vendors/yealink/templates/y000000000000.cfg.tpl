#!version:1.0.0.1
#---------------------------------------PROVISIONING PARAMS------------------------
auto_provision.pnp_enable = 0
auto_provision.dhcp_enable = 1

#Set the auto provisioning mode (0-Disabled (default), 1-Power on, 4-Repeatedly,
#5-Weekly, Power on + Repeatedly, Power on + Weekly)

auto_provision.mode = 4
auto_provision.power_on = 1
auto_provision.repeat.enable = 1
auto_provision.repeat.minutes = 30
auto_provision.server.url = {{domain.http_prov_server}}
auto_provision.schedule.periodic_minute = 1
auto_provision.schedule.time_from = 00:00
auto_provision.schedule.time_to = 00:00
auto_provision.schedule.dayofweek = 0123456

#auto_provision.server.username =
#auto_provision.server.password =
#auto_provision.weekly.enalbe = 0
#auto_provision.weekly.mask = 0123456
#auto_provision.weekly.begin_time = 00:00
#auto_provision.weekly.end_time = 00:00

#Set the AES key used for decrypting the Common CFG file
#auto_provision.aes_key_16.com =

#---------------------------------------SECURITY PARAMS------------------------
security.user_name.admin = admin
security.user_name.user = user
security.user_name.var = var
security.user_password = admin:sip123
security.user_password = user:user1
security.user_password = var:var1

account.1.sip_server.1.register_on_enable =1
account.1.sip_server.1.address = {{ domain.sip_server_ip }}

account.2.sip_server.1.register_on_enable =0
account.2.sip_server.1.address = {{ domain.sip_server2_ip }}

#---------------------------------------NETWORK PARAMS------------------------
#Configure the WAN port type; 0-DHCP(default), 1-PPPoE, 2-Static IP Address #Require reboot
network.internet_port.type = 0
# TODO network.phone_port_vlan={{domain.vlan_voip}}
# TODO network.pc_port_vlan={{domain.vlan_pc}}

#Configure the PC port type;0-Router,1-Bridge(default) #Require reboot
network.bridge_mode = 1
#Set the web server access type (0-Disabled, 1-HTTP&HTTPS (default), 2-HTTP only, #3-HTTPS only)
#Require reboot
network.web_server_type = 1
network.port.http = 80
network.port.https = 443
#---------------------------------------LOCALE PARAMS------------------------
lang.wui = Russian
lang.gui = English

#---------------------------------------FEATURES----------------------------
features.pickup.blf_audio_enable = 1
features.pickup.blf_visual_enable = 1
features.pickup.direct_pickup_code = *8#
features.pickup.direct_pickup_enable = 1
