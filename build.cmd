#!/bin/bash
git pull
git switch develop
docker build ./ -t autorestiot

