package raft

import (
	"reflect"
	"sync"
)

// -------------------------------------------------------------------------------------------------

type reqMsg struct {
	endName  interface{}
	svcMeth  string
	argsType reflect.Type
	args     []byte
	replyCh  chan replyMsg
}

type replyMsg struct {
	ok    bool
	reply []byte
}
type ClientEnd struct {
	endName interface{}   // this is endpoints name
	ch      chan reqMsg   // copy if Network.endCh
	done    chan struct{} //closed when network is cleaning up
}

// -------------------------------------------------------------------------------------------------

// once a new log entry is commited, peer sends an ApplyMsg to the service via applyCh passed to Make().
// CommandValid is set to true to indicate ApplyMsg contains a new log entry
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}
type Raft struct {
	mu        sync.Mutex   // for locking
	peers     []*ClientEnd // Rpc end points of all peers
	persister *Persister   // object to hold this peer's persister state
	me        int          // peers index
	dead      int32        // sent by Kill()
}

// -------------------------------------------------------------------------------------------------
// return currentTerm and whether this server believes its the leader or not
func (rf *Raft) GetState() (int, bool) {
	var term int
	var isLeader bool
	//code
	return term, isLeader
}

type RequestVoteArgs struct {
}

type RequestVoteReply struct {
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
}

// send rpc to the server which is the index in rf.peers[]. The call will always return t/f
// unless there is a problem with the handler function on the server side
// Call() sends a request and waits, if the reply comes with a timeout, its true otherwise Call() returns false

func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

// the service using raft must return the index of the value to be committed
// return false if the server is not the lader otherwise, return the index of where the log would
// have been committed, the currentTerm and true

func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	return index, term, isLeader
}
