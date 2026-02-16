# 6731i.cfg.template: 
# ================================================================
# Date: Sept. 3, 2012
# Companion configuration template: aastra.cfg.template
# This file is only read by 6731i after it has read aastra.cfg

#prgkey1 label:"LogOn"
!prgkey1 type:xml
prgkey1 value:http://$$PROXYURL$$:22222/Logon
prgkey1 states:idle
prgkey2 type:speeddial
#prgkey2 label:"MsgWaiting"
prgkey2 value:"*32#"
#prgkey3 is Diversion key for this model

#prgkey4 available for Individual Key

#In order to prevent prgkey5-8 to be overwritten by individual
#key settings in MX-ONE, these keys are locked.
#SAVE
prgkey5 locked:1
#DELETE
prgkey6 locked:1

prgkey7 type:directory  #if local directory
#If corporate. instead of local directory:
#prgkey7 type:"xml"
#prgkey7 label:"CorpDir"
#prgkey7 value:"http://<CMG Server IP address>/xml/directory/CorpDir.php"
prgkey7 locked:1

prgkey8 type:services
prgkey8 locked:1

