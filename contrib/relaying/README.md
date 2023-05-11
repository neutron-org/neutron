# Relaying

In Cosmos, relayers prove state on chain A to chain B so that assets can safely move betwen chains A and B.  Relayers don't really move tokens -- instead they cryptographically prove that others are allowed to do so. 


## An Optimal Relayer Setup

* Run your own nodes so you control quality
* Use state sync to ensure that there are no security risks from downloading chain state. quicksync.io is in fact not safe.
* Use pebbledb to make your nodes as fast as possible.  Speed matters.
* Run your nodes on the same machine as your relayer, so that there's no latency between your Gaia node, your Neutron node, and your relayer.


## Learning to relay

When you're learning to relay, use the tool `screen` s that you can see log output and know each step of the process. Later, you'll graduate to systemD, Kubernetes, or Docker swarm, and this is right and good and normal.  This guide will use `screen` so that you get a full understanding of the components before trying to automate or orchestrate them. 


