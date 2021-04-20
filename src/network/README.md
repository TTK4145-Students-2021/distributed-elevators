Network module with UDP peer discovery, tcp messages and more
==========================================

This is a collection of modules related to networking. 
The UDP peer discovery and timeout is based on [Network-go](https://github.com/TTK4145/Network-go).
Some inspiration for the messaging server were taken from [jsonpipe](https://github.com/Itoxi-zz/jsonpipe)
[kcp-go](https://pkg.go.dev/github.com/xtaci/kcp-go) is used for reliable, error checked and ordered delivery of streams over UDP packets. 

Network
--------
Responsible for sending messages between elevators and modules. Messages to be sent by the network module are sent on a single send channel, messages to be received by other modules are sent on seperate receive channels. 
Implements messaging server and client. Messaging is done with kcp-go, a reliable UDP-module. Due to the high packet loss simulation requirement, TCP is a really bad choice due to the aggressive congestion control implemented to save bandwith. kcp-go however is optimized for flow rate and uses a bit more bandwidth, but reaches speeds magnitudes faster than TCP in our case. With TCP, we measured delays well over a minute with 25% chance of packet loss, where kcp got delays of about a seconds. 

The server listens to an available UDP port on the computer, incrementing the chosen default port until an available port is found. Clients are added when they are discovered by peers. Any struct can be sent from an elevator to either the master elevator or all connected elevators. A channel address is specified in the header, signaling which channel the recieving elevator should put the message on. This channel adress allows us to be able to send the same struct to different specific channels if we want. 

network module also handles routing messages locally within one elevator's modules if the receipient is themself. 

peers
--------

Implements UDP peer discovery and connection timeout. The peers ID, IP and TCP port is sent at an interval on UDP broadcast. Any update in peers lost or discovered is signaled on a channel. 



masterselection
--------
Handles master selection. A simplified selection algorithm is implemented, simply choosing the master as the elevator with the lowest id. Assuming that all elevators are discovering each other by peers module, this will function within spec, where all elevators agrees on a master without further communication.

conn
--------
OS-specific UDP-broadcast implementation

localip
--------
Gets IP-address
