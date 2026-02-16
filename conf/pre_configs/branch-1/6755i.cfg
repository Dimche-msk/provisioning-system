# 6755i.cfg.template: 
# ================================================================
# Date: Jan. 10, 2012
# Companion configuration template: aastra.cfg.template
# This file is only read by 6755i after it has read aastra.cfg

prgkey1 type:services
prgkey2 type:directory
prgkey3 type:callers
prgkey4 type:speeddial
#prgkey4 label:"MsgWaiting"
prgkey4 value:"*32#"
#prgkey5 is Diversion key for this model

softkey5 label:"LogOn"
!softkey5 type:xml
softkey5 value:http://$$PROXYURL$$:22222/Logon
softkey5 states:idle
softkey6 type:"xml"
softkey6 label:"CorpDir"
softkey6 value:"http://<CMG Server IP address>/xml/directory/CorpDir.php"

#softkey7 and softkey8 can be defined here as well. 
#First Individial key is stored on softkey9 using default ExtensionKeyOffset:8
#in /etc/opt/eri_sn/ip_telephony.conf