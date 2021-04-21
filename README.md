Elevator Project
================

Summary
-------
Golang implementation of  `n` elevators working in parallel across `m` floors. Mutliple elevators can communicate over a network. UDP network protocol is used, with reliable messaging, and peer discovery with connection timeouts. A master-slave model with a dynamic master elevator selection is used to synchronize order assignment and reassignment between the coupled system of elevators. 

Dependencies
-------
kcp-go is required, and is used for smooth, reliable, error checked and ordered delivery of streams over UDP packets.
kcp can be installed by running `go get github.com/xtaci/kcp-go`
Due to the high packet loss simulation requirement, TCP is a really bad choice due to the aggressive congestion control implemented to save bandwith. kcp-go 
is optimized for flow rate and uses a bit more bandwidth, but reaches speeds magnitudes faster than TCP in our case. With TCP, we measured delays well over a minute with 25% chance of packet loss, where kcp got delays of about a seconds. This is partly due to TCP's doubling of timout for each packet loss, selective retransmission implemented with kcp, and non-concessional flow control possible with kcp. 

Running
-------
An elevator can be started by running src/main.go with the flags -id and the optional -simport. The default simport is 15657. The id is required, and needs to be a unique integer for every elevator. 

Master/Slave
-------
The master elevators extra responsibility is to 
  - collect the state of every elevator
  - collect all orders and order completions from all elevators 
  - calculate the resulting order assignments 
  - send orders and assignment to all elevators.  
Every elevator saves all orders to be done, such that orders are not lost if the master elevator dies. 

A master election is initiated whenever a peer connects, choosing the elevator with the lowest id as master.

Module communication
--------
Communication is performed with go channels. The network module is responsible for sending data to the correct elevator, routing it locally if needed.
![UML SANNTID](https://user-images.githubusercontent.com/47594779/115465919-990f2000-a22f-11eb-84dd-98f8111ba3da.png)

Modules
-------

### Network module with UDP peer discovery, reliable udp messaging and more

This is a collection of modules related to networking. 
The UDP peer discovery and timeout is based on [Network-go](https://github.com/TTK4145/Network-go).
Some inspiration for the messaging server were taken from [jsonpipe](https://github.com/Itoxi-zz/jsonpipe)
[kcp-go](https://pkg.go.dev/github.com/xtaci/kcp-go) is used for reliable, error checked and ordered delivery of streams over UDP packets. 

#### network

Responsible for sending messages between elevators and modules. Messages to be sent by the network module are sent on a single send channel, messages to be received by other modules are sent on seperate receive channels. 
Implements messaging server and client. Messaging is done with kcp-go, a reliable UDP-module. Due to the high packet loss simulation requirement, TCP is a really bad choice due to the aggressive congestion control implemented to save bandwith. kcp-go however is optimized for flow rate and uses a bit more bandwidth, but reaches speeds magnitudes faster than TCP in our case. With TCP, we measured delays well over a minute with 25% chance of packet loss, where kcp got delays of about a seconds. 

The server listens to an available UDP port on the computer, incrementing the chosen default port until an available port is found. Clients are added when they are discovered by peers. Any struct can be sent from an elevator to either the master elevator or all connected elevators. A channel address is specified in the header, signaling which channel the recieving elevator should put the message on. This channel adress allows us to be able to send the same struct to different specific channels if we want. 

network module also handles routing messages locally within one elevator's modules if the receipient is themself. 

#### peers


Implements UDP peer discovery and connection timeout. The peers ID, IP and TCP port is sent at an interval on UDP broadcast. Any update in peers lost or discovered is signaled on a channel. 



#### masterselection

Handles master selection. A simplified selection algorithm is implemented, simply choosing the master as the elevator with the lowest id. Assuming that all elevators are discovering each other by peers module, this will function within spec, where all elevators agrees on a master without further communication.

#### conn

OS-specific UDP-broadcast implementation

#### localip

Gets IP-address

### Master module
Responsible for collecting all orders and order updates from all peers, calculate assignment and reassignment and send this to all peers. Cab orders are also saved for each elevator, and resent to an elevator that reconnects. This module only runs if the elevator is determined to be master by the network module. All non-completed orders are kept even if a master disconnects, or a new master is elected, as long as at least one of the peers is alive. The new master then requests all orders from all peers.   
### Orders module
Responsible for keeping the order list by listening for assigned orders from master, sending order updates to master and suppliyng the controller module with local orders and lights. 
### Controller module
The state machine of the single elevator logic. Local elevator orders and lights to be set are supplied by the orders module. Sensor input is supplied by the hardware module. Motor and light setting is sent to the hardware module. 
### Hardware module
Responsible for direct hardware actions to the elevator box or simulator, such as motor and light setting, and sensor and button readouts. 
