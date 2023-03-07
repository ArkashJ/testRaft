package raft

import "sync"

// once a new log entry is commited, peer sends an ApplyMsg to the service via applyCh passed to Make().
// CommandValid is set to true to indicate ApplyMsg contains a new log entry
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

type ClientEnd struct {
	endName interface{}   // this is endpoints name
	ch      chan reqMsg   // copy if Network.endCh
	done    chan struct{} //closed when network is cleaning up
}

type Raft struct {
	mu        sync.Mutex
	peers     []*ClientEnd
	persister *Persister
	me        int
	dead      int32
}
