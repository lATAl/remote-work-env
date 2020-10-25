#!/bin/bash

current_date=$(./fw_ports.py)
export AUTOSSH_GATETIME=0
autossh -M 0 -o "ServerAliveInterval 30" -o "ServerAliveCountMax 3" -N $current_date
# ssh $current_date
