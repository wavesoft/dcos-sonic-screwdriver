# :wrench: dcos-sonic-screwdriver

> A tool for all the tools!

_DC/OS Sonic Screwdriver_ is a tool that installs to your desktop various utilities, scripts and other tools that can make your life easier when you are deploying or debugging stuff on DC/OS

## Installation

Just install the binary in your system:

### For Mac OSX

```
curl -L -o /usr/local/bin/ss \
  https://github.com/wavesoft/dcos-sonic-screwdriver/releases/download/v0.1.2/sonic-screwdriver.darwin && \
  chmod +x /usr/local/bin/ss 
```

### For Linux

WIP

### For OSX

WIP

## Usage

Use `ss ls` to see what tools are available:

```
~$ ss ls
Available tools in the registry:
 marathon-storage-tool  View and modify Marathon ZK state
 ...
```

Use `ss add` to install the tool and make it available for use:

```
~$ ss add marathon-storage-tool
==>  Add marathon-storage-tool
==>  Pulling mesosphere/marathon-storage-tool:1.4.5
1.4.5: Pulling from mesosphere/marathon-storage-tool
Digest: sha256:3bf6ebf419de2a3bb5b8afe5d64d13c43860cb8cb4e7a94e4db59364f0b88c1d
Status: Image is up to date for mesosphere/marathon-storage-tool:1.4.5
👨🏻‍🚀  marathon-storage-tool/1.4.5 has landed!

~$ marathon-storage-tool --zk://marathon-zk-1:2181/marathon
...
```

Use `ss rm` to remove a tool and wipe it's traces:

```
~$ ss rm marathon-storage-tool
==>  Remove marathon-storage-tool
👨🏻‍🚀  marathon-storage-tool has left the rocket ship!
```

