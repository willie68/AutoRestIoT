#!/bin/bash
git pull
git switch develop
docker build ./ -t mcs/autorestiot:V1
docker run -d --restart always --name autorestiot -p 8544:8443 -p 8191:8080  -v /opt/autorest/data:/data mcs/autorestiot:V1
