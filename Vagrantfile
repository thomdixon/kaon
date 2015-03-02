# -*- mode: ruby -*-
# vi: set ft=ruby :

$script = <<SCRIPT
apt-get update && apt-get install -y golang redis-server git
export GOPATH=/vagrant
cd /vagrant
go get github.com/fzzy/radix/redis
go run /vagrant/src/kaon.go &
echo $! > /var/run/kaon.pid
SCRIPT

VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "puppetlabs/ubuntu-14.04-64-puppet"
  config.vm.network "forwarded_port", guest: 8080, host: 9111
  config.vm.provision "shell", inline: $script
end
