What Is Sia?
============

Sia is a compensation based platform for distributed cloud storage. Anybody can join Sia as a storage host and anybody can rent storage from the network. It's a lot like an AirBnB for cloud storage.

Sia is a robust platorm that stores files across hundreds of machines. If any of these machines fail, corrupt, or disappear, the network will automatically repair the lost files with no break in service and with no damage to the original copies. Even though Sia is fully protected against widespread failure or withdrawal of the hosts, the total volume of redundancy is only around 25% - substantially less expensive than using RAID or a mirrored backup. (this is discussed in depth below)

Sia is cheap. Very cheap. Bulk storage costs (amoratized to include racks, electricity, etc.) around $50 per terabyte, and the average disk lasts around 18 months. The raw cost of hosting data in todays market is less than $3 per terabyte per month. If you add in Sia's redundancy and a profit margin, the cost comes to approximately $5 per terabye per month, or 0.05 cents per gigabyte per month. This is half of the price of Amazon Glacier. (which charges 0.1 cents per gigabyte per month).

Sia is fast. Every file is hosted on hundreds of machines, and downloads happen from each machine simultaneosly. You can access your data immediately and saturate your internet connection. As a service, Sia is comparible to Amazon S3. As an expense, Sia is cheaper than Amazon Glacier. This is possible because of the distributed setup that is used to power Sia.

Sia is elastic, meaning you can rent as much or as little storage as you want, and you pay for exactly what you are renting. You can request more data at any time and have your request processed immediately. You never need to guess whether you need the 100GB package or the 500GB package, instead you rent exactly as much as you use, and you rent more or less as you add and remove files. There is no fee for adjusting how much you are renting.

Sia secure. By default, all data is encrypted on the client machine before being uploaded to the network. The encrypted data is then scrambled and divided into hundreds of pieces and distributed across many datacenters. The only person who will be able to see your data is you.

The Sia Economic Model
======================

Sia compensates hosts using a cryptocurrency. This cryptocurrency will have it's own circulation and value, but will be easily exchangeable for bitcoins and subsequently USD. When the currency launches, 10,000 Siacoins will be mined per day and distributed to the storage hosts according to the volume of storage they offer to the network. The number of coins mined per day will decrease until the 4 year mark, at which point the network will mine coins such that the annual inflation rate is kept permanently at 5%. Sia has no premining.

Storage can only be rented using Siacoins. For people unfamiliar with cryptocurrency, there will be a service that allows one to trade dollars directly for storage, transparently exchanging dollars for siacoins behind the scenes. The siacoins payed out to rent storage will go directly to the hosts that store the files.

This means that Sia hosts have two incomes. The first is freshly minted currency, and the second is from people renting the storage. Depending on the price of the Siacoin vs. the price of storage, freshly minted currency could actually produce a weird dynamic where the rent cost of storage is cheaper than the raw cost of the storage (but the rent + mined results in a profit for the hosts).

Siacoins are supposed to be about the storage that they purchase, not about speculation in the currency. That's one of the reasons that we chose to make the Siacoin permanently inflationary. Speculation is thought to be one of the major destabilizers of the bitcoin price. Instead of investing in Siacoin, clients are recommended to purchase only as many as they need to store their data for a comfortable period of time, buying more regularly.

A client doing this isolates themselves from huge swings in the price. If the price of Siacoin suddenly plummets, you only have a few dollars invested anyway. But even more, a sudden drop in the price of Siacoin is likely to be accompanied by a sudden drop in the price of storage on Sia, which means your Siacoins should stretch nearly as far as they would have anyway.

If there is a sudden explosion in the amount of cheap storage available (perhaps Seagate comes out with a new technology), the value of the Siacoin is likely to drop a lot. But the price per terabyte in siacoins should not change at all. Even though you Siacoins are only worth half of what they used to be worth, they will still store just as much data as they would have prior to losing value. This is why we emphasize buying siacoins for the data that gets stored as opposed to their monetary value. The price of siacoins per terabyte should be much more stable than the cost of siacoins per dollar.

Most interestingly, the Siacoin has an inheret use. There is an explicit reason to use Siacoins, which is to rent storage from Sia. This is a resource that can only be accessed using Siacoins, and this makes Siacoin unique to every other cryptocurrency. It puts a minimum on the value of the siacoin - a siacoin will never be worth less than the storage that you can buy using the siacoin. Additionally, if the value of the siacoin skyrockets, new hosts will be incentivized to join the network, which increases the minimum value of the siacoin.

The Sia Company
===============

There is no corporation yet, however there is going to be a startup behind Sia. We've explored many options for revenue, and for now we've settled on a 1.9% fee on all transactions using Siacoins. These fees will be distributed among the owners of Siastock, a secondary currency that will be built into Sia. Siastock will be fully premined and initially owned by Luke and David, the cofounders of the startup behind Sia. The Siastock will be treated somewhat like equity, and initially will be traded away in return for investments. Siastock is fully speculative, and derives its value from the value and trade volume of siacoin.

This reveue will be initially for feeding and housing the developers. As Sia grows, the money will be used for security audits, used to fund support teams, and overall used to build a powerful decentralized ecosystem around Sia. As revenue grows, it will be used to pay open source developers to integrate increasingly large parts of our lives with Sia and cryptocurrencies in general.

If 1.9% seems extreme, I would like to remind you that Sia should be used to buy storage. A 1.9% premium for a market-maker is very small, especially when compared to AirBnb, Uber, and Kickstarter.

The Sia Technology
==================

Sia works quite a bit differently than existing cryptocurrencies. Instead of being proof of work, Sia is proof of storage. We enforce storage proofs by breaking hosts into sets of 128, each member of the set hosting exactly 16GB. If your machine is offering more than 16GB, you participate in multiple of these sets, called 'quorums'.

The quorums individually follow a solution to the byzantine generals problem presented in section 4 of the paper 'The Byzantine Generals Problem'. This provides a solution that guarantees all honest participants in a quorum will revieve the information. Though the solution is complex, we've optimized the networking such that each block only takes a few minutes, comparable to bitcoin's block rate. It's possible that a financial set of blocks could operate at a much faster rate.

Quorums communicate through a tree. Quorums use a verification algorithm to confirm with high probability (99.999999999%) that all blocks they receive are honest. The big advantage is that each quorum can operate knowing mostly about what happens only within that quorum (along with some aggregate information about the branches of the tree they are in - but this is much more lightweight). With existing cryptocurrencies, every miner must know about every transaction, which is expensive and limits the types and volume of transactions that can occur. The exact algorithms for the tree will be discussed at greater length in the yet-incomplete whitepaper.

Finally, Sia will support a scripting system that is much lighter and more powerful than the scripting in existing cryptocurrencies. It will enable much more powerful scripts, and potentially even enable fully decentralized dynamic web hosting. Though the scripting is much more powerful than other cryptocurrencies, it will still be weak and expensive compared to centralized cloud computing.

There is a general assumption that on the network, less than 50% of the hosts will be dishonest, and that no quorum will have more than 80% dishonest hosts. If part 1 is true, part 2 can be achieved by randomly placing hosts around the network. Sia's method of random number generation is a complex topic that will be covered at-length in the whitepaper.

Cryptography and Erasure Coding
===============================

For cryptography, Sia uses libsodium. This includes local random number generation, hashing, and signatures using ECC Curve 25519. As far as we can tell, this is a secure library that follows the best practices. We are not experts in this area however. We want Sia to employ the absolute best practicies when using cryptography, so if you see something that is amiss, please let us know.

For erasure coding, Sia uses a Cauchy Reed Solomon coding library called longhair.

Project Progress
================

Currently, we have most parts of the basic quorum implemented. We have no part of the tree implemented, nor any part of the Byzantine-tolerant DHT that will allow quorums to communicate. We have the erasure coding libraries established and most of the crypto libraries in place (no encryption yet).

We expect to have a basic demo complete by April 25th, but only of the quorums. We expect to have a beta with financial trading, the quorum tree, and some scriping in place by the end of June. We expect to be launching the full system sometime in August.

Where Can I Learn More?
=======================

Primary Contact: David Vorick, david.vorick@gmail.com  
Secondary Contact: Luke Champine, luke.champine@gmail.com  
Mailing List: sia-dev@googlegroups.com  
