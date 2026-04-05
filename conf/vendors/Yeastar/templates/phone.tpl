<?xml version="1.0" encoding="utf-8"?>
<ta>
<system>
<lansettings>
  <dhcp>yes</dhcp>
  <hostname>TA3200</hostname>
  <pppoe>no</pppoe>
</lansettings>
<autopsettings>
  <enablepnp>no</enablepnp>
  <enabledhcp>yes</enabledhcp>
  <autoupflag>0</autoupflag>
  <autoupintertime>180</autoupintertime>
</autopsettings>
</system>
<gateway>
<fxsport itemcount="{{ account.lines|length }}">
{% for line in account.lines %}
  <fxsport{{ forloop.Counter0 }}>
    <idx>{{ forloop.Counter }}</idx>
    <modulestatus>up</modulestatus>
    <channel>{{ forloop.Counter }}</channel>
    <name>{{ line.account_number }}</name>
    <rxflash>200</rxflash>
    <maxcallduration>6000</maxcallduration>
    <callwaiting>no</callwaiting>
    <dnd>no</dnd>
    <ringout>60</ringout>
    <alwaysforward>no</alwaysforward>
    <noanswerforward>yes</noanswerforward>
    <busyforward>yes</busyforward>
    <forwarddesttype>number</forwarddesttype>
    <forwarddest></forwarddest>
    <rxgain>40</rxgain>
    <txgain>50</txgain>
    <voiptemplate>1</voiptemplate>
    <failovervoiptemplate></failovervoiptemplate>
    <registerusername>{{ line.auth_name }}</registerusername>
    <registerauthname>{{ line.auth_name }}</registerauthname>
    <registerfromuser>{{ line.auth_name }}</registerfromuser>
    <registeronlinenumber></registeronlinenumber>
    <registerauthpassword>{{ variables.PasswdPre }}{{ line.account_number }}{{ variables.PasswdPost }}</registerauthpassword>
    <isfax>yes</isfax>
    <forwardmoh>none</forwardmoh>
    <dialpatternidx>1</dialpatternidx>
    <enableforwardprompt>no</enableforwardprompt>
    <didnumber>{{ line.auth_name }}</didnumber>
    <enablehotline>no</enablehotline>
    <hotlinedialdelay>2</hotlinedialdelay>
    <mwisendtype>disable</mwisendtype>
    <cidsignalling>bell</cidsignalling>
    <ciddtmflength>120</ciddtmflength>
    <ciddtmfinterval>120</ciddtmfinterval>
    <answeronpolarityswitch>no</answeronpolarityswitch>
    <hanguponpolarityswitch>no</hanguponpolarityswitch>
    <enablemwi>no</enablemwi>
    <minrxflash>100</minrxflash>
    <cidmode>default</cidmode>
    <cidsendingmode>ring</cidsendingmode>
    <cidtype>bell</cidtype>
    <enablemeteringpluse>no</enablemeteringpluse>
    <meteringfreqency>12khz</meteringfreqency>
    <cadenceactivetime>2000</cadenceactivetime>
    <cadenceinactivetime>2000</cadenceinactivetime>
    <meteringamplitude>500</meteringamplitude>
    <number>{{ line.auth_name }}</number>
    <enableechotraining>yes</enableechotraining>
    <sendflash>no</sendflash>
    <dtmfpassthrough>no</dtmfpassthrough>
    <enableloop>no</enableloop>
    <loopduration>800</loopduration>
  </fxsport{{ forloop.Counter0 }}>
{% endfor %}
</fxsport>
<voiptrunktemplate itemcount="1">
  <voiptrunktemplate0>
    <idx>1</idx>
    <type>SIP</type>
    <templatename>Trunk-{{ domain_name }}</templatename>
    <domain>{{ variables.sip_server | default: domain_name }}</domain>
    <host>{{ variables.sip_server | default: domain_name }}</host>
    <port>5060</port>
    <transport>udp</transport>
    <enableregister>portregister</enableregister>
    <enablesrtp>no</enablesrtp>
    <enableproxyserver>no</enableproxyserver>
    <proxyserver_host></proxyserver_host>
    <proxyserver_port>5060</proxyserver_port>
    <callerid></callerid>
    <codec1>alaw</codec1>
    <codec2>ulaw</codec2>
    <codec3>g729</codec3>
    <codec4></codec4>
    <codec5></codec5>
    <maxchannel>0</maxchannel>
    <qualify>yes</qualify>
    <dtmfmode>rfc2833</dtmfmode>
    <krdiversion></krdiversion>
    <realm></realm>
    <authinvite>no</authinvite>
    <failoverhost></failoverhost>
    <failoverport>5060</failoverport>
    <registerusername></registerusername>
    <registerauthname></registerauthname>
    <registerfromuser></registerfromuser>
    <registeronlinenumber></registeronlinenumber>
    <registerauthpassword></registerauthpassword>
    <qualifyfreq>10</qualifyfreq>
    <enablefailoverproxyserver>no</enablefailoverproxyserver>
    <failoverproxyserver_port>5060</failoverproxyserver_port>
    <randport>no</randport>
    <flashsignal>hf</flashsignal>
    <ignoresdpversion>no</ignoresdpversion>
  </voiptrunktemplate0>
</voiptrunktemplate>
<dialpatterntemplate itemcount="1">
  <dialpatterntemplate0>
    <idx>1</idx>
    <templatename>DialPatternTemplate1</templatename>
    <dialpatterns>.,,</dialpatterns>
  </dialpatterntemplate0>
</dialpatterntemplate>
</gateway>
</ta>
