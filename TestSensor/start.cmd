@echo off
mode con: cols=160 lines=10 
echo starting mosqitto
test-mqtt.exe -b tcp://127.0.0.1:1883 -c wohnzimmer -t stat/temperatur/wohnzimmer -u temp -p temp