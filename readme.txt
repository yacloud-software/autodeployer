# Auto Deployer

The autodeployer and deploymonkey suite are simple tools to
start binaries on a farm of servers.
The deploymonkey keeps track of versions and instructs autodeployer
to start/stop binaries. (or binaries within tar files).
It's design goals are
* simplicity
* no dependencies on build chains
* do not fiddle with networks or containers


deploymonkey is deprecated and replaced by deployminator
=======================================================
a) the name "deploymonkey" is not great - naming is hard
b) it grew and was a patchwork of things
the deployminator is written from ground up to:
* be simple for clients (e.g. buildrepo)
* based on a repeatable algorithm (to 'redistribute')
* handle long downloads on autodeployer gracefully
* provide good feedback to users (developers)
* add deployment strategies easier, e.g. one trigger for
  the rewrite was the need to deploy one-instance-on-each-machine


Resource Limits
===========================
Deploymonkey uses default limits. They may be overriden like so:

namespace: timeseries
groups:
 - groupid: testing
   applications:
    - repository: timeseries
      limits:
       maxmemory: 10000

"maxmemory" is in "megabytes"
It is enforced like so:
mm.Cur=maxmemory*1024*124
mm.Max=mm.Cur
Setrlimit(RLIMIT_AS,&mm)

in bash, that is equivalent to ulimit -v maxmemory*1024
(bash/ulimit is in KBytes)


Deployment with static port
===========================

autodeployer can automatically use DNAT to redirect traffic to new instances:
configure /etc/cnw/autodeployer/config.yaml
example:

applications:
- label: "lbproxy"
  matcher:
    repository: lbproxy
    binary: lbproxy-server
    groupname: testing
    namespace: lbproxy
  actionports:
  - portindex: 1
    publicport: 80
  - portindex: 2
    publicport: 443
- label: ""
  matcher:
    repository: testrepo2
    binary: testbinary2
    groupname: mygroup
    namespace: namespace2
  actionports:
  - portindex: 1
    publicport: 10
  - portindex: 2
    publicport: 22

# nftables...


**** important note ***
https://wiki.nftables.org/wiki-nftables/index.php/Performing_Network_Address_Translation_(NAT)#Masquerading
states this: "with kernel versions before 4.18, you have to register the prerouting/postrouting chains even if you have no rules there"
Tests confirm that below works with 4.19 but not 4.9 kernel

nft add table nat
nft add chain nat prerouting '{ type nat hook prerouting priority 0 ; }'
nft add rule nat prerouting ' tcp dport 26 dnat to :25 ; '

#for autodeployer:
nft add table autodeployer_table
nft add chain autodeployer_table autodeployer '{ type nat hook prerouting priority 0 ; }'
nft add rule autodeployer_table autodeployer ' tcp dport 26 dnat to :25 ; '

# display current rules:
nft -nn list table autodeployer_table 

#which gives us:
table ip autodeployer_table {
        chain autodeployer {
                type nat hook prerouting priority 0; policy accept;
                tcp dport 26 dnat to :25
                tcp dport 27 dnat to :25
        }
}

# the rules may be flushed with:
nft flush chain autodeployer_table autodeployer

# the chain may be read with:
nft -f [filename]
# (this is accumulative - chain won't be automatically flushed prior to reading)

# to integrate into an existing nftables configuration:
# the nft chains install themselves with prerouting and input priority hooks

# relevant packets will be marked with '412'
# so you should add something like this:
nft add rule ip filter INPUT meta mark 412 accept


============================
a version:
a versionID matches the ID of a row in table group_version
appgroup.id==group_version.group_id
lnk_app_grp.group_version_id = ${versionID}
lnk_app_grp.app_id = appdef_store.ByID()
lnk_app_grp.group_version_id = group_version.ID

