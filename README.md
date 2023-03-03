1) Raft is a distributed consensus algorithm where multiple nodes have to agree with each other.

2) Each node is a:
    - Follower
    - Candidate
    - Leader

3) In the beginning every node is a follower and there is no leader.
    Then there is an election where one leader is elected (150ms to 300ms)
    There is an election timer that expires, and a follower becomes a candidate, votes for itself and asks for a vote
    from every other node. Then each node, if they havent voted in the election cycle, vote again. The node then resets the timer.

4) Once the leader is elected, they send AppendEntries messages to its followers.

5) Once a leader is elected, all changes go through the leader. Changes are added as an entry to the log of the leader however until a majority is reached, they are not implemented. Once there is consensus, the leader notifies that the entry has been committed. Entry is committed and node state is updated.

6) The AppendEntries are sent as heartbeats

7) Election term continues until followers stop recieving heartbeats and becomes a candidate

8) There is a case for a split vote (lets say there are 2 clients and 5 nodes). There will be 3 that recieve a majority and the network barrier is broken. 

9) For log replication, after the leader recieved a majority from the replies of the AppendEntries, it commits the changes


-----------------------------------------------------------------------------------------------------------

Leader Election
 
- 2 timeout settings
    election timeout is the time a follower waits before becoming a candidate (rand b/w 150ms and 300ms)
    Node votes for itself and asks for votes (Request Vote)
    if requested node hasnt voted in this term, it votes again and node resets the election timeout

    Leader is elected from the majority and sends an AppendEntries message to its followers. These messages are sent as heartbeats

    Followers respond to each respondEntries message. Election term till followers stops recieveing heartbeats. Then stoo leader and re-election

- In case of split vote, two nodes send messages at the same time. Each candidate has 2 votes and no more for this term. Nodes wait for a new election
-----------------------------------------------------------------------------------------------------------

Log Replication

- Request sent by leader for AppendEntries. Client sends the change to the leader, which commits after a majority is recieved, response sent to the client