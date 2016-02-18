This directory contains samples packer templates to get you started.

* CentOS v7.1
* Debian v8
* OpenSuSE v13.2
* Ubuntu v14.04-3 LTS

All of these templates require that you supply values for the
following variables.

* cid (Client ID)
* cst (Client Secret)
* tid (Client Tenant)
* sid (Subscription ID)
* rgn (Resource Group Name)
* sa  (Storage Account)

```batch
packer build^
  -var cid="593c4dc4-9cd7-49af-9fe0-1ea5055ac1e4"^
  -var cst="GbzJfsfrVkqL/TLfZY8TXA=="^
  -var sid="ce323e74-56fc-4bd6-aa18-83b6dc262748"^
  -var tid="da3847b4-8e69-40bd-a2c2-41da6982c5e2"^
  -var rgn="My Resource Group"^
  -var sa="mystorageaccount"^
  c:\packer\ubuntu_14.04.3-LTS.json
```

All sample templates with the exception of Ubuntu set the
configuration parameter
[ssh_pty](https://www.packer.io/docs/templates/communicator.html) to
true.  These OSs require a pty when executing ssh commands, and
require a password be supplied to elevate via sudo.  (The
execute_command for these templates inject the password to sudo via
[STDIN](https://www.packer.io/docs/provisioners/shell.html).)  The
password **must** be supplied by the user.  A sample script for these
templates is shown below.  Note that the *ssh_pass* parameter is now
supplied.

```batch
packer build^
  -var cid="593c4dc4-9cd7-49af-9fe0-1ea5055ac1e4"^
  -var cst="GbzJfsfrVkqL/TLfZY8TXA=="^
  -var sid="ce323e74-56fc-4bd6-aa18-83b6dc262748"^
  -var tid="da3847b4-8e69-40bd-a2c2-41da6982c5e2"^
  -var rgn="My Resource Group"^
  -var sa="mystorageaccount"^
  -var ssh_pass="packer"^
  c:\packer\CentOS_7.1.json
```

## UNIX OS Generalization

The ARM builders executes OS generalization via a packer provisioner.
The last command executed by a provisioner should be the following.

```sh
/usr/sbin/waagent -force -deprovision+user && export HISTSIZE=0 && sync
```

This ensures that...

1. All SSH host key pairs will be deleted.
1. Cached DHCP leases will be deleted.
1. Nameserver configuration in /etc/resolv.conf (or where appropriate)
   will be deleted.
1. The user name provisioned by packer will be deleted.
1. Root password will be disabled (as appropriate).

OS Generalization used to be an explicit step in the old (SMAPI)
builder, and was hidden from end users.  It has been made into an
explicit step to better support the various UNIX flavors.
