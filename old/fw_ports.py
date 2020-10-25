#!/usr/bin/env python

import yaml
import os
import sys
import subprocess
import time

list_ports = []

with open('hosts') as f:
    host = f.readline()
ssh_host = "{}@{}".format(os.environ['USER'], host.strip())

def parse_ports_from_docker_compose(compose_file):
    global list_ports
    with open(compose_file) as f:
        data = yaml.load(f, Loader=yaml.FullLoader)
        for key, value in data['services'].iteritems():
            for key2, value2 in value.iteritems():
                if key2 == 'ports':
                    list_ports = list_ports + value2
        # print(data['services'])

project_names = [line.rstrip('\n') for line in open('project_name')]
for project_name in project_names:
    if project_name == "pancake_v2":
        list_ports += ["4000"]
    if project_name == "pancake_v2":
        list_ports += ["4003"]
    if project_name == "pancake-v2-client":
        list_ports += ["3000"]
    compose_file = "../{}/docker-compose.yml".format(project_name)
    if os.path.isfile(compose_file):
        parse_ports_from_docker_compose(compose_file)

arg_line = ["-R 9999:127.0.0.1:9999"]
for port in list_ports:
    port = port.split(":")[0]
    ports = port.split("-")
    if len(ports) == 2:
        for port in list(range(int(ports[0]), int(ports[1])+1)):
            arg_line += ["-L {}:127.0.0.1:{}".format(port, port)]
    else:
        arg_line += ["-L {}:127.0.0.1:{}".format(port, port)]
socket_file = "/tmp/ssh/dev-server-socket"
arg_line += ["-S {}".format(socket_file)]
arg_line += [ssh_host]
print " ".join(arg_line)
# command_enter_fw_port = ["autossh", "-M", "-S", socket_file, "-fNT"] + arg_line
# command_exit_fw_port = ["ssh", "-S", "/tmp/ssh/dev-server-socket", "-O", "exit", ssh_host]
# is_connected = False
# def run_command(command):
#     print(command)
#     p = subprocess.Popen(command, stdin=subprocess.PIPE, stdout=subprocess.PIPE)
#     stdout, stderr = p.communicate()
# def fw_port_over_ssh():
#     print "Enter fw port"
#     global is_connected
#     run_command(command_enter_fw_port)
#     is_connected = True
# def exit_port_over_ssh():
#     print "Exit fw port"
#     global is_connected
#     run_command(command_exit_fw_port)
#     is_connected = False
# def check_vpn_then_fw_port_over_ssh():
#     global is_connected
#     vpn_is_connected = False
#     vpn_is_connected = vpnIsConnected()
#     if vpn_is_connected:
#         is_connected or fw_port_over_ssh()
#     else:
#         is_connected and exit_port_over_ssh()
#     time.sleep(1)
#
# if os.path.exists(socket_file):
#     exit_port_over_ssh()
# while 1:
#     check_vpn_then_fw_port_over_ssh()
