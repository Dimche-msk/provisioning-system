#!version:1.0.0.1
## The header above must appear as-is in the first line


##[$MODEL]include:config <xxx.cfg>
##[$MODEL,$MODEL]include:config "xxx.cfg"  


include:config <y000000000000.cfg>

#[T46S, T42S, T41S]
#include:config <T4S-series.cfg>

#[T30P, T31P]
#include:config <T30P.cfg>

#Если параметр в .cfg пустой или закомментирован, телефон вернет его к заводским настройкам (по умолчанию включено).     
overwrite_mode = 1

#Включает «режим исключения». Телефон попытается скачать файл для своей модели; если его нет на сервере, он скачает общий файл (по умолчанию выключено).
specific_model.excluded_mode=0