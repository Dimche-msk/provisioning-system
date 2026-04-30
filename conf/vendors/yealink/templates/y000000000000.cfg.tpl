#!version:1.0.0.1
#---------------------------------------PROVISIONING PARAMS------------------------
auto_provision.pnp_enable = 0
auto_provision.dhcp_enable = 1

#Set the auto provisioning mode (0-Disabled (default), 1-Power on, 4-Repeatedly,
#5-Weekly, Power on + Repeatedly, Power on + Weekly)

auto_provision.mode = 4
auto_provision.power_on = 1
auto_provision.repeat.enable = 1
auto_provision.repeat.minutes = 10
auto_provision.server.url = {{variables.http_prov_server}}:{{variables.http_prov_server_port}}
auto_provision.schedule.periodic_minute = 1
auto_provision.schedule.time_from = 00:00
auto_provision.schedule.time_to = 23:59
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
account.1.sip_server.1.address = {{ variables.sip_server1_url }}

account.1.codec.pcma.enable = 1
account.1.codec.pcmu.enable = 1
account.1.codec.g729.enable = 1
account.1.codec.g722.enable = 1
account.1.codec.g722.priority = 1
account.1.codec.pcma.priority = 2
account.1.codec.g729.priority = 3
account.1.codec.pcma.priority = 4

account.2.sip_server.1.register_on_enable =0
account.2.sip_server.1.address = {{ variables.sip_server2_ip }}

#---------------------------------------LOCALE PARAMS------------------------
lang.wui = Russian
lang.gui = English

#---------------------------------------FEATURES----------------------------
features.pickup.blf_audio_enable = 0
features.pickup.blf_visual_enable = 0
features.pickup.direct_pickup_code = *8#
features.pickup.direct_pickup_enable = 1
########################################################################################
##                                 Time                                              ##       
#######################################################################################
##It configures the time zone.For more available time zones, refer to Time Zones on page 215.
##The default value is +8.
local_time.time_zone = +3

##It configures the time zone name.For more available time zone names, refer to Time Zones on page 215.
##The default time zone name is China(Beijing).
local_time.time_zone_name = Europe/Moscow

local_time.ntp_server1 = {{variables.ntp_server1}}
local_time.ntp_server2 = {{variables.ntp_server2}}

##It configures the update interval (in seconds) when using the NTP server.
##The default value is 1000.Integer from 15 to 86400
#local_time.interval = 

##It enables or disables daylight saving time (DST) feature.
##0-Disabled,1-Enabled,2-Automatic.
##The default value is 2.
local_time.summer_time = 0 

##It configures the way DST works when DST feature is enabled.
##0-DST By Date ,1-DST By Week.
##The default value is 0.
#local_time.dst_time_type = 

##It configures the start time of the DST.
##Value formats are:Month/Day/Hour (for By Date),Month/ Day of Week/ Day of Week Last in Month/ Hour of Day (for By Week)
##The default value is 1/1/0.
#local_time.start_time = 

##It configures the end time of the DST.
##Value formats are:Month/Day/Hour (for By Date),Month/ Day of Week/ Day of Week Last in Month/ Hour of Day (for By Week)
##The default value is 12/31/23.
#local_time.end_time = 

##It configures the offset time (in minutes).
##The default value is blank.Integer from -300 to 300
#local_time.offset_time = 

##It configures the time format.0-12 Hour,1-24 Hour.
##The default value is 1.
local_time.time_format = 1
#0-WWW MMM DD
#1-DD-MMM-YY
#2-YYYY-MM-DD
#3-DD/MM/YYYY
#4-MM/DD/YY
#5-DD MMM YYYY
#6-WWW DD MMM
local_time.date_format = 1  

##It enables or disables the phone to update time with the offset time obtained from the DHCP server.
##It is only available to offset from GMT 0.0-Disabled,1-Enabled.
##The default value is 0.
#local_time.dhcp_time = 

##It configures the phone to obtain time from NTP server or manual settings.0-Manual,1-NTP
##The default value is 1.
#local_time.manual_time_enable = 

##It enables or disables the phone to use manually configured NTP server preferentially.
##0-Disabled (use the NTP server obtained by DHCP preferentially),1-Enabled.
##The default value is 0.
local_time.manual_ntp_srv_prior = 1

#auto_dst.url =

#######################################################################################
##                                    Network CDP                                    ##       
#######################################################################################
static.network.cdp.enable = 0
static.network.cdp.packet_interval = 0



#######################################################################################
##                                    Network IPv6                                   ##       
#######################################################################################
static.network.ipv6_static_dns_enable = 0
static.network.ipv6_icmp_v6.enable = 0
#static.network.ipv6_secondary_dns =
#static.network.ipv6_primary_dns =
#static.network.ipv6_internet_port.gateway =
#static.network.ipv6_internet_port.ip =
#static.network.ipv6_internet_port.type =
#static.network.ipv6_prefix =



#######################################################################################
##                                    Network WiFi                                   ##       
#######################################################################################
##static.wifi.X.label=
##static.wifi.X.ssid=
##static.wifi.X.priority=
##static.wifi.X.security_mode=
##static.wifi.X.cipher_type=
##static.wifi.X.password=
##static.wifi.X.eap_type=
##static.wifi.X.eap_user_name=
##static.wifi.x.eap_password=
##(X ranges from 1 to 5)
##Only T54S/T52S/T48G/T48S/T46G/T46S/T42S/T41S/T29G/T27G Models support these parameters.

#static.wifi.enable =
#static.wifi.1.label =
#static.wifi.1.ssid =
#static.wifi.1.priority =
#static.wifi.1.security_mode =
#static.wifi.1.cipher_type =
#static.wifi.1.password =
#static.wifi.1.eap_type =
#static.wifi.1.eap_user_name =
#static.wifi.1.eap_password =
#static.wifi.show_scan_prompt =
#
###V83 Add
#static.wifi.function.enable =

#######################################################################################
##                                 Network Internet                                  ##       
#######################################################################################
#static.network.ip_address_mode =
#static.network.span_to_pc_port =
#static.network.vlan.pc_port_mode =
#static.network.static_dns_enable =
#static.network.pc_port.enable =
#static.network.primary_dns =
#static.network.secondary_dns =

#Configure the WAN port type; 0-DHCP(default), 1-PPPoE, 2-Static IP Address #Require reboot
#static.network.internet_port.type = 0
#static.network.internet_port.gateway =
#static.network.internet_port.mask =
#static.network.internet_port.ip = {{variables.ip_address}}
# TODO network.phone_port_vlan={{domain.vlan_voip}}
# TODO network.pc_port_vlan={{domain.vlan_pc}}
##V83 Add
#static.network.preference =

#---------------------------------------NETWORK PARAMS------------------------
#Configure the WAN port type; 0-DHCP(default), 1-PPPoE, 2-Static IP Address #Require reboot
#network.internet_port.type = 0
# TODO network.phone_port_vlan={{domain.vlan_voip}}
# TODO network.pc_port_vlan={{domain.vlan_pc}}

#Configure the PC port type;0-Router,1-Bridge(default) #Require reboot
network.bridge_mode = 1
#Set the web server access type (0-Disabled, 1-HTTP&HTTPS (default), 2-HTTP only, #3-HTTPS only)
#Require reboot
network.web_server_type = 1
network.port.http = 80
network.port.https = 443

#######################################################################################
##                               Network Advanced                                    ##       
#######################################################################################
#static.network.dhcp_host_name =
#static.network.dhcp.option60type =
#static.network.mtu_value =
#static.network.qos.audiotos =
#static.network.port.min_rtpport =
#static.network.port.max_rtpport =
#static.network.qos.signaltos =
#
#static.wui.http_enable =
#static.wui.https_enable =
#static.network.port.https =
#static.network.port.http =
#
#static.network.pc_port.speed_duplex =
#static.network.internet_port.speed_duplex =
#
###V83 Add
#static.network.redundancy.mode =
#static.network.redundancy.failback.timeout =



#######################################################################################
##                                   Network LLDP                                    ##       
#######################################################################################
static.network.lldp.enable = 1
#static.network.lldp.packet_interval =



#######################################################################################
##                                   Network VLAN                                    ##       
#######################################################################################
#static.network.vlan.dhcp_enable =
#static.network.vlan.dhcp_option =
#static.network.vlan.vlan_change.enable =
#
#
#static.network.vlan.pc_port_priority =
#static.network.vlan.pc_port_vid =
#static.network.vlan.pc_port_enable =
#
#static.network.vlan.internet_port_priority =
#static.network.vlan.internet_port_vid =
#static.network.vlan.internet_port_enable =


#######################################################################################
##                                    Network VPN                                    ##       
#######################################################################################
#static.network.vpn_enable =
#static.openvpn.url =



#######################################################################################
##                                 Network 802.1x                                    ##       
#######################################################################################
#static.network.802_1x.mode =
#static.network.802_1x.identity =
#static.network.802_1x.md5_password =
#static.network.802_1x.client_cert_url =
#static.network.802_1x.root_cert_url =
#static.network.802_1x.eap_fast_provision_mode =
#static.network.802_1x.anonymous_identity =
#static.network.802_1x.proxy_eap_logoff.enable =
#
#
#static.auto_provision.custom.protect =
#static.auto_provision.custom.sync =
#static.auto_provision.custom.sync.path =
#static.auto_provision.custom.upload_method =




#######################################################################################
##                                    ZERO Touch                                     ##       
#######################################################################################
#static.zero_touch.enable =
#static.zero_touch.wait_time =
#static.features.hide_zero_touch_url.enable =
#static.zero_touch.network_fail_delay_times =
#static.zero_touch.network_fail_wait_times =


#######################################################################################
##                                   Autop URL                                       ##       
#######################################################################################
#static.auto_provision.server.url =
#static.auto_provision.server.username =
#static.auto_provision.server.password =


#######################################################################################
##                                   Autop Weekly                                    ##       
#######################################################################################
#static.auto_provision.weekly.enable =
#static.auto_provision.weekly.dayofweek =
#static.auto_provision.weekly.end_time =
#static.auto_provision.weekly.begin_time =
#static.auto_provision.weekly_upgrade_interval =

#######################################################################################
##                                   Autop Repeat                                    ##       
#######################################################################################
#static.auto_provision.repeat.enable =
#static.auto_provision.repeat.minutes =

#######################################################################################
##                                   Autop DHCP                                      ##       
#######################################################################################
#static.auto_provision.dhcp_option.list_user_options =
#static.auto_provision.dhcp_option.enable =

##V83 Add
#static.auto_provision.dhcp_option.list_user6_options =

#######################################################################################
##                                   Autop Mode                                      ##       
#######################################################################################
#static.auto_provision.power_on =



#######################################################################################
##                               Flexible Autop                                      ##       
#######################################################################################
#static.auto_provision.flexible.end_time =
#static.auto_provision.flexible.begin_time =
#static.auto_provision.flexible.interval =
#static.auto_provision.flexible.enable =

#######################################################################################
##                                 Autoprovision  Other                              ##       
#######################################################################################
#static.auto_provision.prompt.enable =
#static.auto_provision.attempt_expired_time =
#static.auto_provision.attempt_before_failed =
#static.network.attempt_expired_time =
#static.auto_provision.update_file_mode =
#static.auto_provision.retry_delay_after_file_transfer_failed=
#static.auto_provision.inactivity_time_expire =
#static.auto_provision.dns_resolv_timeout =
#static.auto_provision.dns_resolv_nretry =
#static.auto_provision.dns_resolv_nosys =
#static.auto_provision.user_agent_mac.enable =
#static.auto_provision.server.type =
#features.action_uri_force_autop =
#static.auto_provision.url_wildcard.pn =
#static.auto_provision.reboot_force.enable =
#static.auto_provision.dhcp_option.option60_value =
#static.custom_mac_cfg.url =
#static.auto_provision.aes_key_in_file =
#static.auto_provision.aes_key_16.mac =
#static.auto_provision.aes_key_16.com =
#features.custom_version_info =
##V83 Add
#static.auto_provision.authentication.expired_time =
#static.auto_provision.connect.keep_alive =

#######################################################################################
##                                   Autop PNP                                       ##       
#######################################################################################
#static.auto_provision.pnp_enable =



#######################################################################################
##                                  Autop Code                                       ##       
#######################################################################################
##static.autoprovision.X.name
##static.autoprovision.X.code
##static.autoprovision.X.url
##static.autoprovision.X.user
##static.autoprovision.X.password
##static.autoprovision.X.com_aes
##static.autoprovision.X.mac_aes
##Autop Code(X ranges from 1 to 50)

#static.autoprovision.1.name =
#static.autoprovision.1.code =
#static.autoprovision.1.url =
#static.autoprovision.1.user =
#static.autoprovision.1.password =
#static.autoprovision.1.com_aes =
#static.autoprovision.1.mac_aes =



#######################################################################################
##                                   TR069                                           ##       
#######################################################################################

#static.managementserver.enable =
#static.managementserver.username =
#static.managementserver.password =
#static.managementserver.url =
#static.managementserver.periodic_inform_enable =
#static.managementserver.periodic_inform_interval =
#static.managementserver.connection_request_password =
#static.managementserver.connection_request_username =




#######################################################################################
##                                Redirect                                           ##       
#######################################################################################
#static.redirect.user_name =
#static.redirect.password =


#######################################################################################
##                            Firmware Update                                        ##       
#######################################################################################
#static.firmware.url =


#######################################################################################
##                            Confguration                                           ##       
#######################################################################################
features.reset_by_long_press_enable = 1
#features.factory_pwd_enable =
#static.configuration.url =
#static.features.custom_factory_config.enable =
#static.custom_factory_configuration.url =


#######################################################################################
##                               SYSLOG                                              ##       
#######################################################################################
#static.syslog.enable =
#static.syslog.server =
#static.syslog.level =
#static.syslog.server_port =
#static.syslog.transport_type =
#static.syslog.facility =
#static.syslog.prepend_mac_address.enable =
#static.local_log.enable =
#static.local_log.level =
#static.local_log.max_file_size =



#######################################################################################
##                               Log Backup                                          ##       
#######################################################################################
#static.auto_provision.local_log.backup.enable =
#static.auto_provision.local_log.backup.path =
#static.auto_provision.local_log.backup.upload_period =
#static.auto_provision.local_log.backup.append =
#static.auto_provision.local_log.backup.bootlog.upload_wait_time=
#static.auto_provision.local_log.backup.append.max_file_size =
#static.auto_provision.local_log.backup.append.limit_mode=



#######################################################################################
##                                   User Mode                                       ##       
#######################################################################################
#static.security.var_enable = 
#static.web_item_level.url =


#######################################################################################
##                                  Quick Login                                      ##       
#######################################################################################
#wui.quick_login = 


#######################################################################################
##                               Security                                            ##       
#######################################################################################
#static.phone_setting.reserve_certs_enable = 
#features.relog_offtime = 
#static.security.default_ssl_method = 
#static.security.cn_validation = 
#static.security.dev_cert = 
#static.security.ca_cert = 
#static.security.trust_certificates = 
#static.security.user_password = 
#static.security.user_name.var = 
#static.security.user_name.admin = 
#static.security.user_name.user = 

##V83 Add
#static.security.default_access_level =
#phone_setting.reserve_certs_config.enable =


#######################################################################################
##                               Watch Dog                                           ##       
#######################################################################################
#static.watch_dog.enable =

#######################################################################################
##                                   Server Certificates                             ##       
#######################################################################################
#static.server_certificates.url =
#static.server_certificates.delete = 

#######################################################################################
##                           Trusted Certificates                                    ##       
#######################################################################################
#static.trusted_certificates.url =
#static.trusted_certificates.delete =



#######################################################################################
##                           Secure Domain List                                      ##       
#######################################################################################
#wui.secure_domain_list = 


#######################################################################################
##                               Encryption                                          ##       
#######################################################################################
#static.auto_provision.encryption.directory = 
#static.auto_provision.encryption.call_log = 
#static.auto_provision.encryption.config = 




#######################################################################################
##                                   Trnasfer                                        ##       
#######################################################################################
#dialplan.transfer.mode =
#transfer.on_hook_trans_enable =
#transfer.tran_others_after_conf_enable =
#transfer.blind_tran_on_hook_enable =
#transfer.semi_attend_tran_enable =
#phone_setting.call_appearance.transfer_via_new_linekey=


#######################################################################################
##                                   Conference                                      ##       
#######################################################################################
#features.conference.with_previous_call.enable =
#features.local_conf.combine_with_one_press.enable=
#phone_setting.call_appearance.conference_via_new_linekey=



#######################################################################################
##                                   Anonymous                                       ##       
#######################################################################################
#features.anonymous_response_code=



#######################################################################################
##                          Call Configuration                                       ##       
#######################################################################################
#phone_setting.incoming_call_when_dialing.priority=
#phone_setting.hold_or_swap.mode=
#features.play_held_tone.interval=
#features.play_held_tone.delay=
#features.play_held_tone.enable=
#features.play_hold_tone.interval=
#features.ignore_incoming_call.enable=
#force.voice.ring_vol=
#features.mute.autoanswer_mute.enable=
#features.play_hold_tone.delay =
#phone_setting.end_call_net_disconnect.enable =
#features.custom_auto_answer_tone.enable=
#default_input_method.dialing=
#features.speaker_mode.enable=
#features.headset_mode.enable=
#features.handset_mode.enable=
#features.conference.local.enable =
#features.off_hook_answer.enable=
#features.caller_name_type_on_dialing=
#phone_setting.show_code403=
#phone_setting.ring_for_tranfailed=
#features.password_dial.length=
#features.password_dial.prefix=
#features.password_dial.enable=
#features.group_listen_in_talking_enable=
#phone_setting.call_info_display_method=
#phone_setting.called_party_info_display.enable =
#features.headset_training=
#features.headset_prior=
#features.dtmf.replace_tran =
#features.dtmf.transfer =
#phone_setting.ringing_timeout=
#phone_setting.ringback_timeout=
#transfer.multi_call_trans_enable =
#features.keep_mute.enable=
#linekey.1.shortlabel=
#features.config_dsskey_length.shorten =
#transfer.dsskey_deal_type =
#features.auto_linekeys.enable=
#phone_setting.call_appearance.calls_per_linekey=
#features.linekey_call_with_default_account=
###V83 Add
#features.station_name.value =
#features.station_name.scrolling_display =
#voice.headset.autoreset_spk_vol =
#voice.handset.autoreset_spk_vol =
#voice.handfree.autoreset_spk_vol =
#features.headset.ctrl_call.enable = 
#phone_setting.incoming_call.reject.enable =


#######################################################################################
##                           Custom Softkey                                          ##       
#######################################################################################
#phone_setting.custom_softkey_enable=
#custom_softkey_talking.url=
#custom_softkey_ring_back.url=
#custom_softkey_dialing.url=
#custom_softkey_connecting.url=
#custom_softkey_call_in.url=
#custom_softkey_call_failed.url=
#
###V83 Add
#features.homescreen_softkey.acd.enable =
#features.homescreen_softkey.hoteling.enable =
#phone_setting.custom_softkey.apply_to_states =
#features.custom_softkey_dynamic.enable =


#######################################################################################
##                                   Features Bluetooth                              ##       
#######################################################################################
##Only T54S/T52S/T48G/T48S/T46G/T46S/T42S/T41S/T29G/T27G Models support the parameter.
#features.bluetooth_enable=
#features.bluetooth_adapter_name=
##V83 Add
#static.bluetooth.function.enable =


#######################################################################################
##                                  Features USB Record                              ##       
#######################################################################################
##Only T54S/T52S/T48G/T48S/T46G/T46S/T42S/T41S/T29G/T27G Models support the parameter.
features.usb_call_recording.enable =1

#######################################################################################
##                                  Features USB                                     ##       
#######################################################################################
##V83 Add
static.usb.power.enable = 1

#######################################################################################
##                                    Codec                                          ##       
#######################################################################################
#voice.g726.aal2.enable=


#######################################################################################
##                                   DTMF                                            ##       
#######################################################################################
#features.dtmf.min_interval=
#features.dtmf.volume=
#features.dtmf.duration =

#######################################################################################
##                                   Tones                                           ##       
#######################################################################################
#voice.tone.autoanswer =
#voice.tone.message =
#voice.tone.stutter =
#voice.tone.info =
#voice.tone.dialrecall =
#voice.tone.callwaiting =
#voice.tone.congestion =
#voice.tone.busy =
#voice.tone.ring =
#voice.tone.dial =
voice.tone.country = Russia
#voice.side_tone =
#features.partition_tone =
#voice.tone.secondary_dial=

#######################################################################################
##                                   Jitter Buffer                                   ##       
#######################################################################################
#voice.jib.normal=
#voice.jib.max =
#voice.jib.min =
#voice.jib.adaptive =
#
#voice.jib.wifi.normal=
#voice.jib.wifi.max=
#voice.jib.wifi.min=
#voice.jib.wifi.adaptive=

#######################################################################################
##                                   Echo Cancellation                               ##       
#######################################################################################
voice.echo_cancellation =1
#voice.cng =
#voice.vad =

################################################################
#                        SIP Backup Server                    ##
################################################################
#static.network.dns.ttl_enable =
#static.network.dns.last_cache_expired.enable=
#static.network.dns.last_cache_expired
#static.network.dns.query_timeout =
#static.network.dns.retry_times =
#sip.dns_transport_type=
#sip.skip_redundant_failover_addr=


################################################################
#                        SIP Basic Config                     ##
################################################################
#sip.use_out_bound_in_dialog=
#sip.unreg_with_socket_close=
#phone_setting.disable_account_without_username.enable=
#features.auto_answer.first_call_only=

################################################################
#                        SIP Advanced config                  ##
################################################################
#sip.request_validation.event=
#sip.sdp_early_answer_or_offer=
#sip.cid_source.preference=
#sip.request_validation.digest.realm=
#sip.request_validation.digest.list=
#sip.request_validation.source.list=
#sip.send_keepalive_by_socket=
#sip.reliable_protocol.timerae.enable=
#sip.requesturi.e164.addglobalprefix=
#sip.trust_ctrl=
#sip.mac_in_ua=
#
#sip.timer_t1=
#sip.timer_t2=
#sip.timer_t4=
#
#sip.listen_mode=
#sip.listen_port=
#sip.tls_listen_port=
#sip.tcp_port_random_mode=
#sip.escape_characters.enable=
#sip.notify_reboot_enable=
#sip.send_response_by_request=
#sip.disp_incall_to_info=
#features.call_invite_format=
#phone_setting.early_media.rtp_sniffer.timeout=
#sip.reg_surge_prevention =
#
###V83 Add
#sip.dhcp.option120.mode =

################################################################
#                        NAT&ICE                              ##
################################################################
#static.sip.nat_turn.enable=
#static.sip.nat_turn.username=
#static.sip.nat_turn.password=
#static.sip.nat_turn.server=
#static.sip.nat_turn.port=
#
#static.sip.nat_stun.enable=
#static.sip.nat_stun.server=
#static.sip.nat_stun.port=
#
#
#static.ice.enable=
#static.network.static_nat.enable=
#static.network.static_nat.addr=

#######################################################################################
##                           DNS                                                     ##       
#######################################################################################
#dns_cache_a.1.name =
#dns_cache_a.1.ip = 
#dns_cache_a.1.ttl =
#dns_cache_srv.1.name = 
#dns_cache_srv.1.port =
#dns_cache_srv.1.priority =
#dns_cache_srv.1.target =
#dns_cache_srv.1.weight =
#dns_cache_srv.1.ttl =
#dns_cache_naptr.1.name =
#dns_cache_naptr.1.order =
#dns_cache_naptr.1.preference =
#dns_cache_naptr.1.replace =
#dns_cache_naptr.1.service = 
#dns_cache_naptr.1.ttl = 

#######################################################################################
##                                 RTP                                               ##
#######################################################################################
features.rtp_symmetric.enable=


#######################################################################################
##                                 RTCP-XR                                           ##
#######################################################################################
#voice.rtcp.enable=
#voice.rtcp_cname=
#voice.rtcp_xr.enable=
#phone_setting.vq_rtcpxr_display_symm_oneway_delay.enable=
#phone_setting.vq_rtcpxr_display_round_trip_delay.enable=
#phone_setting.vq_rtcpxr_display_moscq.enable=
#phone_setting.vq_rtcpxr_display_moslq.enable = 
#phone_setting.vq_rtcpxr_display_packets_lost.enable=
#phone_setting.vq_rtcpxr_display_jitter_buffer_max.enable=
#phone_setting.vq_rtcpxr_display_jitter.enable=
#phone_setting.vq_rtcpxr_display_remote_codec.enable=
#phone_setting.vq_rtcpxr_display_local_codec.enable=
#phone_setting.vq_rtcpxr_display_remote_call_id.enable=
#phone_setting.vq_rtcpxr_display_local_call_id.enable=
#phone_setting.vq_rtcpxr_display_stop_time.enable=
#phone_setting.vq_rtcpxr_display_start_time.enable=
#phone_setting.vq_rtcpxr_interval_period=
#phone_setting.vq_rtcpxr_delay_threshold_critical=
#phone_setting.vq_rtcpxr_delay_threshold_warning=
#phone_setting.vq_rtcpxr_moslq_threshold_critical=
#phone_setting.vq_rtcpxr_moslq_threshold_warning=
#phone_setting.vq_rtcpxr.interval_report.enable=
#phone_setting.vq_rtcpxr.states_show_on_gui.enable=
#phone_setting.vq_rtcpxr.states_show_on_web.enable=
#phone_setting.vq_rtcpxr.session_report.enable=


#######################################################################################
##                                   Contact                                         ##       
#######################################################################################
#static.directory_setting.url=
#super_search.url=
#
#local_contact.data.url=
#local_contact.data.delete=
#
###Only T54S/T52S/T48G/T48S/T46G/T46S/T29G Models support the parameter
#phone_setting.contact_photo_display.enable=
#
#phone_setting.incoming_call.horizontal_roll_interval=
#
###Only T54S/T52S/T48G/T48S/T46G/T46S/T29G Models support the parameter
#local_contact.data_photo_tar.url=
#local_contact.photo.url=
#local_contact.image.url=
#
###Only T48G/S Models support the parameter
#local_contact.icon_image.url=
#local_contact.icon.url=


#######################################################################################
##                                 Remote Phonebook                                  ##       
#######################################################################################
##remote_phonebook.data.X.url
##remote_phonebook.data.X.name 
##(X ranges from 1 to 5)

#remote_phonebook.data.1.url=
#remote_phonebook.data.1.name=
#features.remote_phonebook.enter_update_enable=
#features.remote_phonebook.flash_time=
#features.remote_phonebook.enable=
#remote_phonebook.display_name=
#
#directory_setting.remote_phone_book.enable =
#directory_setting.remote_phone_book.priority =



#######################################################################################
##                                 LDAP                                              ##       
#######################################################################################
#ldap.enable=
#ldap.user=
#ldap.password=
#ldap.base=
#ldap.port=
#ldap.host=
#ldap.customize_label=
#ldap.incoming_call_special_search.enable=
#ldap.tls_mode=
#ldap.search_type=
#ldap.numb_display_mode=
#ldap.ldap_sort=
#ldap.call_in_lookup=
#ldap.version =
#ldap.display_name=
#ldap.numb_attr=
#ldap.name_attr=
#ldap.max_hits=
#ldap.number_filter=
#ldap.name_filter=
#ldap.call_out_lookup=
#directory_setting.ldap.enable =
#directory_setting.ldap.priority =


#######################################################################################
##                                 History                                           ##       
#######################################################################################
#static.auto_provision.local_calllog.write_delay.terminated=
#static.auto_provision.local_calllog.backup.path=
#static.auto_provision.local_calllog.backup.enable=
#super_search.recent_call=
#features.call_out_history_by_off_hook.enable=
#features.save_call_history=
#features.call_log_show_num=
#search_in_dialing.history.enable=
#search_in_dialing.history.priority=
#directory_setting.history.enable=
#directory_setting.history.priority
#features.save_init_num_to_history.enable=
#features.redial_via_local_sip_server.enable=

##V83 Add
#features.calllog_detailed_information =


#######################################################################################
##                          Contact Backup                                           ##       
#######################################################################################
#static.auto_provision.local_contact.backup.path =
#static.auto_provision.local_contact.backup.enable=


#######################################################################################
##                          Contact Other                                            ##       
#######################################################################################
#directory.search_type=
#directory_setting.local_directory.enable =
#directory_setting.local_directory.priority =

##V83 Add
#phone_setting.search.highlight_keywords.enable =

#######################################################################################
##                          Favorites                                                ##       
#######################################################################################
##V83 Add
local_contact.favorite.enable =
phone_setting.favorite_sequence_type =

#######################################################################################  
##                                  Programablekey                                   ##  
####################################################################################### 
#programablekey.X.type
#programablekey.X.line
#programablekey.X.value
#programablekey.X.xml_phonebook
#programablekey.X.history_type
#programablekey.X.label(X ranges from 1 to 4)
#programablekey.X.extension
##Programablekey X ranges(T48G/T48S/T46G/T46S: X=1-10, 12-14;T42G/T42S/T41P/T41S/T40P/T40G: X=1-10, 13;T29G/T27P/T27G: X=1-14;T23P/T23G/T21(P) E2: 1-10, 14;T19(P) E2: X=1-9, 13, 14;)##


#programablekey.1.type =
#programablekey.1.label =
#programablekey.1.value =
#programablekey.1.line =
#programablekey.1.history_type =
#programablekey.1.xml_phonebook =
#programablekey.1.extension =
#
###V83 Add
#programablekey.type_range.custom =

#######################################################################################  
##                                  Linekey                                          ##  
####################################################################################### 
##linekey.X.line
##linekey.X.value
##linekey.X.extension
##linekey.X.type
##linekey.X.xml_phonebook
##linekey.X.shortlabel
##linekey.X.label
##LineKeyX ranges(T48G/S: X ranges from 1 to 29. T54S/T46G/T46S/T29G: X ranges from 1 to 27. T42G/T42S/T41P/T41S: X ranges from 1 to 15. T40P/T40G/T23P/T23G: X ranges from 1 to 3. T52S/T27P/T27G: X ranges from 1 to 21. T21(P) E2: X ranges from 1 to 2.)##
## Not support T19P_E2

#linekey.1.label =
#linekey.1.line =
#linekey.1.value =
#linekey.1.extension =
#linekey.1.type =
#linekey.1.xml_phonebook =

##V83 Add
#linekey.type_range.custom =


#######################################################################################  
##                                  Dsskey                                           ##  
####################################################################################### 
#features.block_linekey_in_menu.enable =
#features.shorten_linekey_label.enable =
#features.flash_url_dsskey_led.enable =
#features.config_dsskey_length =
#phone_setting.page_tip =
#features.keep_switch_page_key.enable=
#
###phone_setting.idle_dsskey_and_title.transparency(Only support T54S/T52S/T48G/T48S)
#phone_setting.idle_dsskey_and_title.transparency=
#
###V83 Add
#phone_setting.keytype_sequence =
#phone_setting.dsskey_label.display_method =
#local.dsskey_type_config.mode =


#######################################################################################  
##                                Expansion Key                                      ##  
####################################################################################### 
##expansion_module.X.key.Y.type
##expansion_module.X.key.Y.line
##expansion_module.X.key.Y.value
##expansion_module.X.key.Y.extension
##expansion_module.X.key.Y.label
##expansion_module.X.key.Y.xml_phonebook
## Expansion Key X ranges(SIP-T54S/T52S: X ranges from 1 to 3, Y ranges from 1 to 60; SIP-T48G/T48S/T46G/T46S:X ranges from 1 to 6, Y ranges from 1 to 40; SIP-T29G/T27P/T27G:X ranges from 1 to 6, Y ranges from 1 to 20, 22 to 40 (Ext key 21 cannot be configured).)##
## Only SIP-T54S/T52S/T48G/T48S/T46G/T46S/T29G/T27P/T27G Models support the parameter.

#expansion_module.1.key.1.type =
#expansion_module.1.key.1.label =
#expansion_module.1.key.1.value =
#expansion_module.1.key.1.line =
#expansion_module.1.key.1.extension =
#expansion_module.1.key.1.xml_phonebook =
#expansion_module.page_tip.blf_call_in.led =
#expansion_module.page_tip.blf_call_in.enable =
#
###V83 Add
#expkey.type_range.custom =


#######################################################################################  
##                                EDK                                                ##  
####################################################################################### 
##EDK Soft Keys(X ranges from 1 to 10)

#features.enhanced_dss_keys.enable=
#edk.id_mode.enable=
#softkey.1.position=
#softkey.1.use.dialing=
#softkey.1.softkey_id=
#softkey.1.use.dialtone=
#softkey.1.use.conferenced=
#softkey.1.use.held=
#softkey.1.use.hold=
#softkey.1.use.transfer_ring_back=
#softkey.1.use.ring_back=
#softkey.1.use.call_failed=
#softkey.1.use.on_talk=
#softkey.1.use.transfer_connecting=
#softkey.1.use.connecting=
#softkey.1.use.incoming_call=
#softkey.1.use.idle=
#softkey.1.action=
#softkey.1.label=
#softkey.1.enable=
#edk.edklist.1.action=
#edk.edklist.1.mname=
#edk.edklist.1.enable=
#edk.edkprompt.1.enable=
#edk.edkprompt.1.label=
#edk.edkprompt.1.type=
#edk.edkprompt.1.userfeedback=



#######################################################################################  
##                                XML                                                ##  
####################################################################################### 
#push_xml.server=
#push_xml.sip_notify=
#push_xml.block_in_calling=
#default_input_method.xml_browser_input_screen=
#
###V83 Add
#hoteling.authentication_mode =
#push_xml.phonebook.search.delay =
#features.xml_browser.loading_tip.delay =
#features.xml_browser.pwd =
#features.xml_browser.user_name =
#push_xml.password =
#push_xml.username =


#######################################################################################  
##                                  Forward                                          ##  
#######################################################################################  
#features.fwd.allow=
#features.fwd_mode=
#forward.no_answer.enable=
#forward.busy.enable=
#forward.always.enable=
#forward.no_answer.timeout=
#forward.no_answer.on_code=
#forward.no_answer.off_code=
#forward.busy.off_code=
#forward.busy.on_code=
#forward.always.off_code=
#forward.always.on_code=
#forward.no_answer.target=
#forward.busy.target=
#forward.always.target=
#
#features.forward.emergency.authorized_number=
#features.forward.emergency.enable=
#forward.idle_access_always_fwd.enable=
#features.forward_call_popup.enable=
#
###V83 Add
#features.forward.no_answer.show_ring_times =


#######################################################################################  
##                                  DND                                              ##  
#######################################################################################
features.dnd.allow= 0
#features.dnd_mode=
features.dnd.enable= 0

#features.dnd.off_code=
#features.dnd.on_code=

#features.dnd.emergency_authorized_number=
#features.dnd.emergency_enable=
#features.dnd.large_icon.enable=

##V83 Add
#features.keep_dnd.enable =

#######################################################################################  
##                               Phone Lock                                          ##  
#######################################################################################
phone_setting.phone_lock.enable=
phone_setting.phone_lock.lock_key_type=
phone_setting.phone_lock.unlock_pin=
phone_setting.emergency.number=
phone_setting.phone_lock.lock_time_out=



#######################################################################################  
##                               Hotdesking                                          ##  
#######################################################################################
phone_setting.logon_wizard=
phone_setting.logon_wizard_forever_wait=

hotdesking.startup_register_name_enable=
hotdesking.startup_username_enable=
hotdesking.startup_password_enable=
hotdesking.startup_sip_server_enable=
hotdesking.startup_outbound_enable=

hotdesking.dsskey_register_name_enable=
hotdesking.dsskey_username_enable=
hotdesking.dsskey_password_enable=
hotdesking.dsskey_sip_server_enable=
hotdesking.dsskey_outbound_enable=


#######################################################################################  
##                               Voice Mail                                          ##  
#######################################################################################
features.voice_mail_alert.enable=
features.voice_mail_popup.enable=
features.voice_mail_tone_enable=
features.hide_feature_access_codes.enable=



#######################################################################################  
##                               Text Message                                        ##  
#######################################################################################
features.text_message.enable=
features.text_message_popup.enable=





#######################################################################################  
##                               Audio Intercom                                      ##  
#######################################################################################
features.intercom.mode=
features.intercom.subscribe.enable=
features.intercom.led.enable=
features.intercom.feature_access_code=
features.blf.intercom_mode.enable=
features.intercom.ptt_mode.enable=

features.redial_tone=
features.key_tone=
features.send_key_tone=

features.intercom.allow=
features.intercom.barge=
features.intercom.tone=
features.intercom.mute=


voice.handset_send=
voice.handfree_send =
voice.headset_send =
features.intercom.headset_prior.enable=
features.ringer_device.is_use_headset=
features.intercom.barge_in_dialing.enable=



#######################################################################################  
##                               Feature General                                     ##  
#######################################################################################
features.ip_call.auto_answer.enable=
features.show_default_account=
features.call.dialtone_time_out=
features.missed_call_popup.enable=
features.auto_answer_tone.enable=
features.play_hold_tone.enable=
features.key_as_send=
features.send_pound_key=
features.busy_tone_delay=
features.hotline_delay=
features.hotline_number=
features.direct_ip_call_enable=
features.call_num_filter=
features.call_completion_enable=
features.allow_mute=
features.auto_answer_delay=
features.normal_refuse_code=
features.dnd_refuse_code=
features.upload_server=
features.dtmf.repetition=
features.dtmf.hide_delay=
features.dtmf.hide=
features.play_local_dtmf_tone_enable =
features.reboot_in_talk_enable =
features.fwd_diversion_enable=

call_waiting.enable=
call_waiting.tone=
call_waiting.off_code=
call_waiting.on_code=

auto_redial.times=
auto_redial.interval=
auto_redial.enable=


sip.rfc2543_hold=
sip.use_23_as_pound=
forward.international.enable=
phone_setting.headsetkey_mode=
phone_setting.is_deal180=
phone_setting.change_183_to_180=

#######################################################################################  
##                               Action URL&URI                                      ##  
#######################################################################################
#features.csta_control.enable=
#features.action_uri.enable=
#features.action_uri_limit_ip=
#features.show_action_uri_option=
#action_url.call_remote_canceled=
#action_url.remote_busy=
#action_url.cancel_callout=
#action_url.handfree=
#action_url.headset=
#action_url.unheld=
#action_url.held=
#action_url.transfer_failed=
#action_url.transfer_finished=
#action_url.answer_new_incoming_call=
#action_url.reject_incoming_call=
#action_url.forward_incoming_call=
#action_url.ip_change=
#action_url.idle_to_busy=
#action_url.busy_to_idle=
#action_url.call_terminated=
#action_url.missed_call=
#action_url.unmute=
#action_url.mute=
#action_url.unhold=
#action_url.hold=
#action_url.always_fwd_off =
#action_url.always_fwd_on =
#action_url.attended_transfer_call =
#action_url.blind_transfer_call =
#action_url.busy_fwd_off =
#action_url.busy_fwd_on =
#action_url.call_established =
#action_url.call_waiting_off =
#action_url.call_waiting_on =
#action_url.dnd_off =
#action_url.dnd_on =
#action_url.incoming_call =
#action_url.no_answer_fwd_off =
#action_url.no_answer_fwd_on =
#action_url.off_hook =
#action_url.on_hook =
#action_url.outgoing_call =
#action_url.register_failed =
#action_url.registered =
#action_url.setup_autop_finish =
#action_url.setup_completed =
#action_url.transfer_call =
#action_url.unregistered =



#######################################################################################  
##                               Power LED                                           ##  
#######################################################################################
phone_setting.hold_and_held_power_led_flash_enable=
phone_setting.mute_power_led_flash_enable=
phone_setting.talk_and_dial_power_led_enable=
phone_setting.mail_power_led_flash_enable=
phone_setting.ring_power_led_flash_enable=
phone_setting.common_power_led_enable=
phone_setting.missed_call_power_led_flash.enable=


#######################################################################################
##                                  Time&Date                                        ##       
#######################################################################################
#lcl.datetime.date.format =
#auto_dst.url =
#local_time.manual_time_enable =
#local_time.manual_ntp_srv_prior =
#local_time.time_format =
#local_time.date_format =
#local_time.dhcp_time =
#
#local_time.summer_time =
#local_time.dst_time_type =
#local_time.start_time =
#local_time.end_time =
#local_time.offset_time =
#local_time.interval =
#
#local_time.ntp_server1 =
#local_time.ntp_server2 =
#local_time.time_zone =
#local_time.time_zone_name =



#######################################################################################
##                           Multicast Paging                                        ##       
#######################################################################################
##multicast.listen_address.X.label 
##multicast.paging_address.X.channel
##multicast.listen_address.X.ip_address 
##multicast.paging_address.X.ip_address
##multicast.paging_address.X.label
##multicast.listen_address.X.channel
##multicast.listen_address.X.volume
##Multicast(X ranges from 1 to 31.)

#multicast.codec=
#
#multicast.paging_address.1.channel=
#multicast.paging_address.1.label=
#multicast.paging_address.1.ip_address=
#multicast.receive_priority.enable=
#multicast.receive_priority.priority=
#
#multicast.receive.use_speaker=
#multicast.receive.enhance_volume=
#multicast.receive.ignore_dnd.priority=
#
#multicast.listen_address.1.channel=
#multicast.listen_address.1.label=
#multicast.listen_address.1.ip_address=
#multicast.listen_address.1.volume=


#######################################################################################
##                           Preference&Status                                       ##       
#######################################################################################
##Not support T19P_E2
#static.features.default_account=

##Logo File Format: .dob
##Resolution: SIP-T42G/T42S/T41P/T41S: <=192*64  2 gray scale;SIP-T27P/G: <=240*120  2 gray scale;SIP-T40P/T40G/T23P/T23G/T21(P) E2/T19(P) E2: <=132*64  2 gray scale##
#phone_setting.lcd_logo.mode=
#lcd_logo.delete=
#lcd_logo.url=
#
#phone_setting.contrast=
#phone_setting.backlight_time=
#phone_setting.inactive_backlight_level=
#phone_setting.active_backlight_level=
#phone_setting.predial_autodial=
#
#ringtone.url=
#ringtone.delete=
#phone_setting.ring_type=
#phone_setting.inter_digit_time=
#
###Only T54S Model supports the parameter
#phone_setting.idle_clock_display.enable =

#######################################################################################
##                           Digitmap                                                ##       
#######################################################################################
#dialplan.digitmap.enable=
#dialplan.digitmap.string=
#dialplan.digitmap.no_match_action=
#dialplan.digitmap.interdigit_short_timer=
#dialplan.digitmap.interdigit_long_timer=
#dialplan.digitmap.apply_to.press_send=
#dialplan.digitmap.apply_to.forward=
#dialplan.digitmap.apply_to.history_dial=
#dialplan.digitmap.apply_to.directory_dial=
#dialplan.digitmap.apply_to.on_hook_dial=
#dialplan.digitmap.active.on_hook_dialing=
#
###V83 Add
#dialplan.digitmap.apply_to.prefix_key =




#######################################################################################
##                           Emergency Dialplan                                      ##       
#######################################################################################
#dialplan.emergency.enable=
#dialplan.emergency.1.value=
#dialplan.emergency.server.1.address=
#dialplan.emergency.server.1.transport_type=
#dialplan.emergency.server.1.port=
#dialplan.emergency.1.server_priority=
#dialplan.emergency.custom_asserted_id=
#dialplan.emergency.asserted_id_source=
#dialplan.emergency.asserted_id.sip_account= 
#dialplan.emergency.held.request_element.1.name=
#dialplan.emergency.held.request_element.1.value= 
#dialplan.emergency.held.request_type=
#dialplan.emergency.held.server_url=



#######################################################################################
##                               Dialplan                                            ##       
#######################################################################################
#dialplan_replace_rule.url=
#dialplan.replace.line_id.1=
#dialplan.replace.replace.1=
#dialplan.replace.prefix.1=
#phone_setting.dialnow_delay=
#dialplan_dialnow.url=
#dialplan.dialnow.line_id.1=
#dialplan.dialnow.rule.1=
#dialplan.block_out.line_id.1=
#dialplan.block_out.number.1=
#dialplan.area_code.line_id =
#dialplan.area_code.max_len =
#dialplan.area_code.min_len=
#dialplan.area_code.code=

#######################################################################################
##                                 Rings Settings                                    ##
#######################################################################################
#distinctive_ring_tones.alert_info.1.ringer=
#distinctive_ring_tones.alert_info.1.text=

#######################################################################################
##                                   IME Settings                                    ##
#######################################################################################
#directory.search_default_input_method=
#directory.edit_default_input_method=
#gui_input_method.url=

##V83 Add
##Only T48G/T48S Models support the parameter
#phone_setting.virtual_keyboard.enable =

#######################################################################################
##                                   Language Settings                               ##
#######################################################################################
#wui_lang.url=
#wui_lang_note.url=
#wui_lang.delete=
#gui_input_method.delete=
#gui_lang.url=
#gui_lang.delete=
#static.lang.gui=
#static.lang.wui=


#######################################################################################
##                                   Screensaver                                     ##
#######################################################################################
#screensaver.type=
#screensaver.delete=
#screensaver.upload_url=
#features.blf_active_backlight.enable=
#screensaver.display_clock.enable=
#screensaver.clock_move_interval=
#screensaver.picture_change_interval=
#screensaver.wait_time=
#screensaver.xml_browser.url=



#######################################################################################
##                                  Power Saving                                     ##
#######################################################################################
features.power_saving.enable= 0
#features.power_saving.power_led_flash.on_time=
#features.power_saving.power_led_flash.off_time=
#features.power_saving.office_hour.monday=
#features.power_saving.office_hour.tuesday=
#features.power_saving.office_hour.wednesday=
#features.power_saving.office_hour.thursday=
#features.power_saving.office_hour.friday=
#features.power_saving.office_hour.saturday=
#features.power_saving.office_hour.sunday =
#features.power_saving.user_input_ext.idle_timeout=
#features.power_saving.off_hour.idle_timeout=
#features.power_saving.office_hour.idle_timeout=
#features.power_saving.intelligent_mode=


#######################################################################################
##                           Backgrounds  Settings                                   ##
#######################################################################################
##File Formate:
##SIP-T54S/T52S/T48S/T48G/T46G/T46S/T29G: .jpg/.png/.bmp/.jpeg;
##Resolution:
##SIP-T48S/T48G:<=2.0 megapixels;
##for SIP-T46G/T46S/T29G: <=1.8 megapixels;SIP-T54S/T52S:<=4.2 megapixels;
##Single File Size: <=5MB 
##2MB of space should bereserved for the phone

#wallpaper_upload.url=
#phone_setting.backgrounds=

## phone_setting.backgrounds_with_dsskey_unfold(Only support T48G/S)
#phone_setting.backgrounds_with_dsskey_unfold=

##expansion_module.backgrounds(Only support T54S/T52S)
#expansion_module.backgrounds=


#######################################################################################
##                               BSFT Setting                                        ##       
#######################################################################################
#bw.enable =


#######################################################################################
##                                  BLF/BLF List                                     ##       
#######################################################################################
#phone_setting.auto_blf_list_enable=
#phone_setting.blf_list_sequence_type=
#
#blf.enhanced.parked.enable=
#blf.enhanced.parked.led =
#blf.enhanced.parked.talking.action =
#blf.enhanced.parked.callin.action =
#blf.enhanced.parked.idle.action =
#
#blf.enhanced.talking.enable=
#blf.enhanced.talking.led=
#blf.enhanced.talking.talking.action =
#blf.enhanced.talking.callin.action =
#blf.enhanced.talking.idle.action =
#
#blf.enhanced.callout.enable =
#blf.enhanced.callout.led=
#blf.enhanced.callout.talking.action =
#blf.enhanced.callout.callin.action =
#blf.enhanced.callout.idle.action =
#
#blf.enhanced.callin.enable =
#blf.enhanced.callin.led=
#blf.enhanced.callin.talking.action =
#blf.enhanced.callin.callin.action=
#blf.enhanced.callin.idle.action=
#
#blf.enhanced.idle.enable=
#blf.enhanced.idle.led=
#blf.enhanced.idle.talking.action=
#blf.enhanced.idle.callin.action=
#blf.enhanced.idle.idle.action=
#
#features.blf_list_version=
#sip.sub_refresh_random=
#sip.terminate_notify_sub_delay_time=
#
#features.blf_led_mode=
#features.blf_pickup_only_send_code=

#######################################################################################
##                                   SCA                                             ##       
#######################################################################################
#features.auto_release_bla_line=
#features.barge_in_via_username.enable=



#######################################################################################
##                                   Call Park                                       ##       
#######################################################################################
#features.call_park.enable=
#features.call_park.park_mode=
#features.call_park.park_code=
#features.call_park.park_retrieve_code=
#features.call_park.direct_send.enable=
#features.call_park.park_visual_notify_enable=
#features.call_park.park_ring=
#features.call_park.group_enable=
#features.call_park.group_park_code=
#sip.call_park_without_blf=
#features.call_park.line_restriction.enable=


#######################################################################################
##                                    Broadsoft ACD                                  ##       
#######################################################################################
#acd.enable=
#acd.auto_available_timer=




#######################################################################################
##                              Broadsoft XSI                                        ##       
#######################################################################################
#bw.xsi.enable=
#sip.authentication_for_xsi =
#default_input_method.xsi_password=


#######################################################################################
##                             Broadsoft Network Directory                           ##       
#######################################################################################
#bw.xsi.directory.enable=
#bw.calllog_and_dir =
#bw.xsi.call_log.enable=
#bw_phonebook.custom=
#bw_phonebook.enterprise_common_enable=
#bw_phonebook.enterprise_common_displayname=
#bw_phonebook.enterprise_enable=
#bw_phonebook.enterprise_displayname=
#bw_phonebook.group_common_enable=
#bw_phonebook.group_common_displayname=
#bw_phonebook.personal_enable=
#bw_phonebook.personal_displayname=
#bw_phonebook.group_enable=
#bw_phonebook.group_displayname =
#directory.update_time_interval=
#bw.xsi.directory.alphabetized_by_lastname.enable=
#directory_setting.bw_directory.enable =
#directory_setting.bw_directory.priority =
#search_in_dialing.bw_directory.enable =
#search_in_dialing.bw_directory.priority =
###V83 Add
#bw.xsi.directory.update.enable =

#######################################################################################
##                             Broadsoft Network Calllog                             ##       
#######################################################################################
##V83 Add
#bw.xsi.call_log.delete.enable =
#bw.xsi.call_log.multiple_accounts.enable =
#phone_setting.ring_duration =
#

#######################################################################################
##                                Call Pickup                                        ##       
#######################################################################################
#features.pickup.direct_pickup_enable =
#features.pickup.group_pickup_enable =
#features.pickup.direct_pickup_code =
#features.pickup.group_pickup_code =
#features.pickup.blf_audio_enable =
#features.pickup.blf_visual_enable =
#features.pickup_display.method =




#######################################################################################
##                                Alert Info                                         ##       
#######################################################################################
#features.alert_info_tone =


#######################################################################################
##                       Broadsoft Visual Voice Mail                                 ##       
#######################################################################################
#bw.voice_mail.visual.enable=
#voice_mail.message_key.mode=
#bw.voice_mail.visual.display_videomail.enable=



#######################################################################################
##                              Broadsoft Call Recording                             ##       
#######################################################################################
#bw.call_recording.mode =


#######################################################################################
##                              Broadsoft Call Decline                               ##       
#######################################################################################
#features.call_decline.enable =


#######################################################################################
##                         BLF Ring Type                                             ##       
#######################################################################################
#features.blf.ring_type =



#######################################################################################
##                         Features Sync                                             ##       
#######################################################################################
#features.feature_key_sync.enable =
#features.forward.feature_key_sync.local_processing.enable =
#features.forward.feature_key_sync.enable =
#features.dnd.feature_key_sync.local_processing.enable =
#features.dnd.feature_key_sync.enable =
#call_waiting.mode =


#######################################################################################
##                           Broadsoft  UC                                           ##       
#######################################################################################
##Only T54S/T52S/T48G/T48S/T46G/T46S/T29G Models support the parameter
#bw.xmpp.enable =
#features.uc_password =
#features.uc_username =
#bw.xmpp.presence_icon.mode =
#bw.xmpp.change_presence.force_manual.enable =
#bw.xmpp.change_presence.enable =
#phone_setting.dsskey_directory_auto.enable =
#features.uc_dir.match_tail_number=
#directory_setting.bw_uc_buddies.enable =
#directory_setting.bw_uc_buddies.priority =
#search_in_dialing.bw_uc_buddies.enable =
#search_in_dialing.bw_uc_buddies.priority =
#
###V83 Add
#phone_setting.uc_favorite_sequence_type =

#######################################################################################
##                      Broadsoft  Emergency Call                                    ##       
#######################################################################################
##V83 Add
bw.emergency_calling.enable  =



#######################################################################################
##                         Metaswitch Setting                                        ##       
#######################################################################################
#meta.enable =
#meta.login_mode =
#meta.comm_portal.server.username =
#meta.comm_portal.server.password =
#meta.comm_portal.server.url =
#meta.comm_portal.enable =
#meta.comm_portal.contacts.update_interval =
#meta.comm_portal.acd.enable=
#meta.comm_portal.replace_local_call_list.enable=
#meta.comm_portal.contacts.group.mlhgs.label=
#meta.comm_portal.contacts.group.extensions.label=
#meta.comm_portal.contacts.group.contacts.label=
#meta.comm_portal.contacts.group.mlhgs.enable=
#meta.comm_portal.contacts.group.extensions.enable=
#meta.comm_portal.contacts.group.contacts.enable=
#meta.comm_portal.call_list.enable=
#meta.comm_portal.contacts.enable=
#meta.comm_portal.message.enable=
#meta.comm_portal.logout.enable =
#meta.comm_portal.keep_alive_interval_time =
#
###V83 Add
#directory_setting.meta_directory.enable=
#directory_setting.meta_directory.priority=
#directory_setting.meta_call_log.enable=
#directory_setting.meta_call_log.priority=
#search_in_dialing.meta_call_log.priority =
#search_in_dialing.meta_call_log.enable =
#search_in_dialing.meta_directory.priority =
#search_in_dialing.meta_directory.enable =



#######################################################################################
##                                   Genbend Setting                                 ##       
#######################################################################################
#gb.sopi.enable=
#gb.sopi.gab.enable=
#gb.sopi.pab.enable=
#features.pab.soupuser=
#features.pab.enable=
#gb.sopi.pab.match_in_calling.enable=
#gb.sopi.gab.retain_search_filter=
#gb.sopi.service_url=
#gb.sopi.password=
#gb.sopi.username=
#directory_setting.gb_gab_directory.priority =
#directory_setting.gb_gab_directory.enable =
#directory_setting.gb_pab_directory.enable =
#directory_setting.gb_pab_directory.priority =
#search_in_dialing.gb_pab_directory.priority =
#search_in_dialing.gb_pab_directory.enable =


#######################################################################################
##                                   Loopback Call                                   ##       
#######################################################################################
##V83 Add
#sip.loopback.enable =
#sip.loopback_type =
#sip.pkt_loopback_mode
#sip.loopback.auto_answer.mode =
#sip.pkt_loopback_encapsulated_payload =
#sip.pkt_loopback_directed_payload =
