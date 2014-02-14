Folders
=======

'common' contains objects and functions that need to be used by all of the other folders. It is a bit like a 'misc', but is explicitly for functions that have broad applications in the program. All of the cryptographic functions are in common, for example.

'disk' is for all functions that interface with the disk. This is for storing and accessing files, which needs to be done frequently for the storage proofs.

'network' is for communicating over the internet. It listens for messages from other hosts and sends messages out.

'swarm' runs the logic for blockchains and the blocktree.

General Flow
============

The daemon starts up and looks at a config file, figures out how many swarms it needs to build and if it needs to load anything from the disk. Also does a bunch of error checking at this stage. Then it builds a set of state objects (one for each swarm the machine is participating in) and gives them to a network object.

The network object starts listening to the network for incoming updates. Each update is sorted and sent to its respective state object. After all the listening is set up, the network object is in a waiting state - either something will come in off the listener object, or one of the state objects will submit something along with a list of where it should be sent.

Objects coming in will either be Blocks, Heartbeats, Updates, or Indictments and each object will be intended for a specific swarm. The network object (tentatively called the multiplexor) will determine what the object is, which swarm it is for, and then it will call a function to have the swarm address the object.

Objects going out will either be Blocks, Heartbeats, or Indictments. The concensus algorithm also has a special outgoing message called an 'Ack' (confirming it recieved a heartbeat) and another one tentatively called a 'Post' (to talk to the block compiler)

Incoming Heartbeats
===================

Incoming heartbeats have to be verified and then just get put into a block. The id of the heartbeat is the id of the host that sent the heartbeat.

Incoming Blocks
===============

Incoming blocks are merely a large list of heartbeats, but they signal that the state is moving forward a step. Once a block is verified as valid, every heartbeat in the block is applied to the state. This means processing all transactions, collectively verifying storage proofs, verifying entropy and creating new random numbers.

Incoming Updates
================

Incoming updates are added to the stash of updates that will be included in the next heartbeat. They get processed when blocks are announced.

Incoming Indictments
====================

Incoming indictments are processed immediately. An indictment does not need concsensus of the swarm to be processed, as indictments contain incriminating data that any single host can verify. Indictments are then passed along to the neighbors, and included in the next block. Note: They are included in the block for historical purposes. Indictments are processed immediately!

Outgoing Heartbeats
===================

Heartbeats need to be sent immidiately after an incoming block has been processed. The contain all updates recieved to the node, entropy, and file proofs.

Outgoing Blocks
===============

You only ever release a block if you are the block compiler. You release a block a set amount of time after the previous block (algorithm for determing that to be decided later). Before releasing the block, you process all of the 'Posts' that you have recieved to figure out which heartbeats belong in the block.

Outgoing Indictments
====================

You only send an indictment if you have evidence that another host has performed an illegal action, such as producing an 'Ack' and then not following through, or approving of an illegal block or transaciton.

Acks
====

An Ack, or acknowledge, is sent after successfully recieving a heartbeat. This is part of the concensus process.

Posts
=====

After a certain amount of time from releasing the previous block, the host sends a 'Post' to the block compiler, who must then return an 'Ack'. The post informs the block compiler of all the heartbeats thus far recieved. This is part of the concensus algorithm for the blockchain.
