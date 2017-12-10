#!/bin/bash

adduser {{.User}} -gecos "" --disabled-password
echo "{{.User}} ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

mkdir /home/{{.User}}/.ssh
echo "{{.Key}}" > /home/{{.User}}/.ssh/authorized_keys