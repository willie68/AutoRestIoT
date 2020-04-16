@echo off
echo starting
test-mqtt.exe -b tcp://127.0.0.1:1883 -c wohnzimmer -t data/temperatur/wohnzimmer -u temp -p temp