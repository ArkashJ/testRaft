package raft

import (
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

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

	electionTimer  *time.Timer
	heartbeatTimer *time.Timer
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
	rf.mu.Lock()
	defer rf.mu.Unlock()

	term = rf.currentTerm
	if rf.serverState == "leader" {
		isLeader = true
	} else {
		isLeader = false
	}

	return term, isLeader
}

// -------------------------------------------------------------------------------------------------
type AppendEntriesArgs struct {
	term         int
	leaderID     string
	prevLogTerm  int
	prevLogIndex int
	entries      []LogItem
	leaderCommit int
}

type AppendEntriesReply struct {
	term    int
	success bool
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	reply.term = rf.currentTerm
	reply.success = false

	if args.term < rf.currentTerm { // case when term lags behind currentTerm
		return
	}

	_, lastLogIndex := rf.getLastLogItem()
	if rf.log[args.prevLogIndex].Term == 0 { // case where log doesnt have an term at prevLogIndex
		return
	}

	if rf.log[args.prevLogIndex].Term != args.prevLogTerm { //case with entry conflict

		remainLen := lastLogIndex - args.prevLogIndex
		//deleting the existing entries and ones that follow it
		for i := args.prevLogIndex; i < args.prevLogIndex+remainLen; i++ {
			rf.log[i].Term = 0
		}
		//setting the prevLogIndex term to prevLogTerm
		rf.log[args.prevLogIndex].Term = args.prevLogTerm
	}

	if lastLogIndex < args.prevLogIndex {
		//appending all entries that are not in the log
		for i := lastLogIndex; i < args.prevLogIndex; i++ {
			rf.log[i].Term = args.entries[i-lastLogIndex].Term
		}
	}

	if args.leaderCommit > rf.commitIndex {
		// Find the minimum of commitIndex and args.prevLogIndex
		rf.commitIndex = int(math.Min(float64(args.leaderCommit), float64(args.prevLogIndex)))
	}

	reply.term = args.term
	reply.success = true
	rf.currentTerm = args.term
}

// -------------------------------------------------------------------------------------------------
// RequestVoteArgs contains the currentTerm of the candidate, its id, index and term of the candidates last log entry
type RequestVoteArgs struct {
	term         int
	candidateId  string
	lastLogIndex int
	lastLogTerm  int
}

// RequestVoteReply will contain the reply of the RequestVote Rpc.
type RequestVoteReply struct {
	term        int
	voteGranted bool
}

// rpc call to request for a vote which sets the term and returns whether the vote was granted or not
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	reply.term = rf.currentTerm
	reply.voteGranted = false

	if args.term < rf.currentTerm {
		return
	}

	if rf.votedFor != "" || rf.votedFor != args.candidateId {
		return
	}

	serverLastLogIndex, severLastLogTerm := rf.getLastLogItem()
	if args.lastLogTerm != severLastLogTerm || (args.lastLogTerm == severLastLogTerm && args.lastLogIndex < serverLastLogIndex) {
		return
	}

	reply.term = args.term
	reply.voteGranted = true
	rf.currentTerm = args.term
	rf.votedFor = args.candidateId
}

// send rpc to the server (index in rf.peers[]). Call always returns t/f unless problem with server side handler function
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	// Call() sends a request and waits, if the reply comes with a timeout, its true otherwise Call() returns false
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
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

// -------------------------------------------------------------------------------------------------
// the code does not halt go routines but calls Kill(), one can check if go routines are killed by looking at the value
// Atomic prevents the need for locking
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
}

func (rf *Raft) Killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

// -------------------------------------------------------------------------------------------------
func (rf *Raft) startElectionTimer() {
	rf.electionTimer = time.NewTimer(5 * time.Second)
}

func (rf *Raft) stopElectionTimer() {
	rf.electionTimer.Stop()
}

func (rf *Raft) resetElectionTimer() {
	rf.stopElectionTimer()
	rf.startElectionTimer()
}

// -------------------------------------------------------------------------------------------------

func (rf *Raft) startNewElection() {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// increment current term, change state to candidate, send RequestVoteRPC and vote for self

	rf.currentTerm++
	rf.votedFor = "" //set in the RequestVote function
	rf.serverState = "candidate"

	lastLogIndex, lastLogTerm := rf.getLastLogItem()

	var voteRecieved int32 = 1

	for server := range rf.peers {
		if server == rf.me {
			continue
		}

		args := &RequestVoteArgs{
			term:         rf.currentTerm,
			candidateId:  strconv.Itoa(rf.me),
			lastLogIndex: lastLogIndex,
			lastLogTerm:  lastLogTerm,
		}
		var reply RequestVoteReply

		go func(server int, args *RequestVoteArgs, reply *RequestVoteReply) {
			if rf.sendRequestVote(server, args, reply) {
				rf.mu.Lock()
				defer rf.mu.Unlock()

				if reply.term > rf.currentTerm {
					rf.currentTerm = reply.term
					rf.votedFor = ""
					rf.serverState = "follower"
				}

				if reply.voteGranted {
					atomic.AddInt32(&voteRecieved, 1)
					if atomic.LoadInt32(&voteRecieved) > int32(len(rf.peers)/2) {
						rf.serverState = "leader"
						rf.nextIndex = make([]int, len(rf.peers))
						rf.matchIndex = make([]int, len(rf.peers))
					}
				}
			}
		}(server, args, &reply)
	}
}

// -------------------------------------------------------------------------------------------------

func (rf *Raft) sendHeartbeats() {

}

// -------------------------------------------------------------------------------------------------
// start a new election if the server hasnt recieved a heartbeat
func (rf *Raft) ticker() {
	for !rf.Killed() {
		//start a new leader election if the peer hasnt heard from the leader in a while
		select {
		case <-time.After(time.Duration(5) * time.Second):
			rf.startNewElection()
			for rf.serverState != "leader" { // deals with the case when there is a stalemate
				rf.startNewElection()
			}
		default:
		}
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
