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

All templates with the exception of Ubuntu set the configuration
parameter
[ssh_pty](https://www.packer.io/docs/templates/communicator.html) to
true.  These OSs require a pty when executing ssh commands, which
generally causes **sudo** commands to fail (YMMV).  Please review the
following issues for details.

* [mitchellh/packer #1804](https://github.com/mitchellh/packer/issues/1804)
* [mitchellh/packer #2420](https://github.com/mitchellh/packer/issues/2420)
* [mitchellh/packer #2423](https://github.com/mitchellh/packer/issues/2423)
* etc.

