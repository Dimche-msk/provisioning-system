#!version:1.0.0.1
auto_provision.pnp_enable = 0
auto_provision.power_on = 1
auto_provision.repeat.enable = 1
auto_provision.repeat.minutes = 30
auto_provision.server.url = http://10.78.50.6

auto_provision.dhcp_enable = 1

security.user_name.admin = admin
security.user_name.user = user
security.user_name.var = var
security.user_password = admin:sip123
security.user_password = user:user1
security.user_password = var:var1

account.1.sip_server.1.register_on_enable =1
account.1.sip_server.1.address = {{ sip_server_ip }}

account.2.sip_server.1.register_on_enable =1
account.2.sip_server.1.address = {{ sip_server_ip }}

