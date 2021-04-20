Elevator Project
================

Summary
-------
Golang implementation of  `n` elevators working in parallel across `m` floors. Mutliple elevators can communicate over a network. UDP network protocol is used, with reliable messaging, and peer discovery with connection timeouts. A master-slave model with a dynamic master elevator selection is used to synchronize order assignment and reassignment between the coupled system of elevators. 

Dependencies
-------
kcp-go is required, and is used for smooth, reliable, error checked and ordered delivery of streams over UDP packets.
kcp can be installed by running ...
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
![UML SANNTID](https://user-images.githubusercontent.com/47594779/115465919-990f2000-a22f-11eb-84dd-98f8111ba3da.png)

