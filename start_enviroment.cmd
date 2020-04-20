@echo off
e:
echo start Mongodb
rem cd \Sprachen\platform\mongodb\
start /min  E:\Sprachen\platform\mongodb\start.cmd
timeout /T 10
echo start mosquitto
cd \Sprachen\platform\mosquitto\
start /min E:\Sprachen\platform\mosquitto\start.cmd
cd \daten\git-sourcen\AutoRestIoT
