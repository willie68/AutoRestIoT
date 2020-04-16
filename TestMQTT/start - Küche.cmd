@echo off
echo starting
test-mqtt.exe -b tcp://127.0.0.1:1883 -c kueche -t data/temperatur/kueche