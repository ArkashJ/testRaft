package raft

import (
	"reflect"
	"sync"
	"sync/atomic"
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

// -------------------------------------------------------------------------------------------------
type Raft struct {
	mu        sync.Mutex   // for locking
	peers     []*ClientEnd // Rpc end points of all peers
	persister *Persister   // object to hold this peer's persister state
	me        int          // peers index
	dead      int32        // sent by Kill()
	//states constant across all servers
	currentTerm int
	votedFor    string
	log         []LogItem
	//states volatile for all servers
	commitIndex int
	lastApplied int
	//states needed by the leader
	nextIndex  []int
	matchIndex []int
	// server can be follower, candidate or leader
	serverState string
}

type LogItem struct {
	Term    int         // store the term
	command interface{} // since raft is agnostic to the commands it recieved, command can be of anytype and thus interface
}

// -------------------------------------------------------------------------------------------------
// helper function to get the last log item
func (rf *Raft) getLastLogItem() (int, int) {
	lastLogIndex := len(rf.log) - 1
	lastLogTerm := rf.log[lastLogIndex].Term

	return lastLogIndex, lastLogTerm
}

// -------------------------------------------------------------------------------------------------
// return currentTerm and whether this server believes its the leader or not
func (rf *Raft) GetState() (int, bool) {
	var term int
	var isLeader bool
	//code
	return term, isLeader
}

// -------------------------------------------------------------------------------------------------
// RequestVoteArgs contains the currentTerm of the candidate, its id, index and term of the candidates last log entry
type RequestVoteArgs struct {
	term         int
	candidateId  string
	lastLogIndex int
	lastLogTerm  string
}

// -------------------------------------------------------------------------------------------------
// RequestVoteReply will contain the reply of the RequestVote Rpc.
type RequestVoteReply struct {
	term        int
	voteGranted bool
}

// -------------------------------------------------------------------------------------------------
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) error {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	reply.term = rf.currentTerm

	if args.term < rf.currentTerm {
		reply.voteGranted = false
	}

	if (rf.votedFor == "" || rf.votedFor == args.candidateId) && args.lastLogTerm == rf.log[args.lastLogIndex] {
		reply.voteGranted = true
	}

	reply.term = args.term
	rf.currentTerm = args.term
	rf.votedFor = args.candidateId

	return nil
}

// -------------------------------------------------------------------------------------------------
// send rpc to the server which is the index in rf.peers[]. The call will always return t/f
// unless there is a problem with the handler function on the server side
// Call() sends a request and waits, if the reply comes with a timeout, its true otherwise Call() returns false
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

// -------------------------------------------------------------------------------------------------
// the service using raft must return the index of the value to be committed
// return false if the server is not the lader otherwise, return the index of where the log would
// have been committed, the currentTerm and true
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	return index, term, isLeader
}

// -------------------------------------------------------------------------------------------------
func (rf *Raft) persist() {
}

// -------------------------------------------------------------------------------------------------
// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
}

// -------------------------------------------------------------------------------------------------
// the code does not halt go routines but calls Kill(), one can check if go routines are killed by looking at the value
// Atomic prevents the need for locking
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
}

// -------------------------------------------------------------------------------------------------
func (rf *Raft) Killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

// -------------------------------------------------------------------------------------------------
// start a new election if the server hasnt recieved a heartbeat
func (rf *Raft) ticker() {
	for !rf.Killed() {
	}
}

// -------------------------------------------------------------------------------------------------
func Make(peers []*ClientEnd, me int, persister *Persister, apply chan ApplyMsg) *Raft {
	rf := &Raft{
		peers:     peers,
		me:        me,
		persister: persister,

		currentTerm: 0,
		votedFor:    "",
		log:         make([]LogItem, 0),

		commitIndex: 0,
		lastApplied: 0,

		nextIndex:  make([]int, 0),
		matchIndex: make([]int, 0),

		serverState: "follower",
	}

	rf.readPersist(persister.ReadRaftState())
	go rf.ticker()

	return rf
}
