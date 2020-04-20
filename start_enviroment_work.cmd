@echo off
e:
echo start Mongodb
cd \SPRACHEN\platform\mongoDb\
start /min  E:\SPRACHEN\platform\mongoDb\start_4.2.cmd
timeout /T 2
echo start mosquitto
cd \Sprachen\platform\mosquitto\
start /min E:\SPRACHEN\platform\mqtt\mosquitto\start.cmd
cd \daten\git-sourcen\AutoRestIoT
