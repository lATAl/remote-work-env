#!/usr/bin/env python

import os
from subprocess import call

with open('/Users/tuan/dev/setup-work-env/hosts') as f:
    host = f.readline()
ssh_host = "{}@{}".format(os.environ['USER'], host.strip())

def parse_ports_from_docker_compose(compose_file):
    list_ports = []
    with open(compose_file) as f:
        data = yaml.load(f, Loader=yaml.FullLoader)
        for key, value in data['services'].iteritems():
            for key2, value2 in value.iteritems():
                if key2 == 'ports':
                    list_ports += value2
    return list_ports

full_path = os.getcwd()
path, filename = os.path.split(full_path)
socket_file = "/tmp/ssh/dev-server-socket"
command = ["mosh", ssh_host,"-- tmux new -A -s {}".format(filename)]
command = " ".join(command)
call(command, shell=True)
