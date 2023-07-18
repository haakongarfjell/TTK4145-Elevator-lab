Elevator Project - TTK4145 Real-time Programming
================

Summary
-------

The project goal was to create software for controlling `n` elevators working in parallel across `m` floors. Our solution is a pure peer-to-peer communication system with UDP broadcast, ensuring all elevators know the full system state at any given time. 


Requirements
-----------------

There were a number of requirements our system needed to fulfill. They are summarized in the following points:
* The button lights are a service guarantee 

* No calls are lost

* The lights and buttons should function as expected

* The door should function as expected

* An individual elevator should behave sensibly and efficiently

The full project description is available [here](https://github.com/TTK4145/Project).

Our solution
----------------------
We chose to implement our system in `golang`. This allowed us to utilize built-in concurrency features such as goroutines and channels, which were a crucial part of our solution.

Our system is based on a pure peer-to-peer design. Using UDP broadcast, all elevators on the network continuously share their internal state and their current knowledge of the active hall-orders with eachother. In addition, all the elevators broadcast their own ID every 15ms to a designated peers-port, allowing each elevator to keep track of who is connected to the network at all times. 

All cab-calls are stored and serviced locally, and if an elevator crashes with active cab-calls it retains them once restarted. Hall-calls are shared between all elevators on the network, and the state of an order is updated according to a cyclic-counter. When a hall-button is pushed, once all the elevators are aware of the new order they all run their local distributor. Since all elevators have full knowledge of the system they distribute the order equally, to the most suitable elevator. When a hall-order has been serviced all elevators are made aware of this, and when the order is cleared the cyclic-counter resets.

The elevators monitor whether they are stuck or not with the help of a stuck-timer. If this timer expires, the elevator labels itself as unavailable. In the event of a disconnect, software/hardware crash or someone being stuck, hall-orders are redistributed between those elevators that are still fully operational. A disconnected elevator is still able to service cab-calls. When an elevator reconnects to the network, the active hall-orders in the system are redistributed to include the connected elevator.

Modules
-------


**orders** <br />
    The central module in our system. It recieves from and sends to the network module, recieves new hall-orders from the local elevator, interacts with the distributor module, and sends distributed orders back to the local elevator. 

**local_elevator** <br />
    The module contains all the types, functions and logic necessary to run a single elevator.
    It consists of six submodules:
  * backup <br />
    Contains functions for writing and reading cab-calls to/from file.
  * elev_io <br />
    A handed-out file containing types and functions necessary for interacting with the elevator hardware.
  * elev_state_machine <br />
    Contains all the functions for state transitions in the local elevator.
  * elev_struct <br />
    Contains types and functions that define a single elevator.
  * requests <br />
    Contains functions for operating on a single elevator's requests.
  * single_elevator <br />
    Contains the go routine to run a single elevator, which is started from main.
    

**network** <br />
    A handed-out network module that can be found [here](https://github.com/TTK4145/Network-go). Contains functionality for UDP communication and monitoring of available peers.

**distributor** <br />
    This module contains the types and function necessary to run the executable `hall_request_assigner`, a handed-out program that optimally distributes orders. Can be found [here](https://github.com/TTK4145/Project-resources/tree/master/cost_fns)

**config** <br />
    This module contains constants used globally in our program.

<br />

Additional resources
--------------------

Go to [the project resources repository](https://github.com/TTK4145/Project-resources) to find more resources for doing the project. 




