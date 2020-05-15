@echo off
docker build ./ -t mcs/autorestiot:V1
docker run --name autorestiot -p 9443:9443 -p 9080:9080 mcs/autorestiot:V1