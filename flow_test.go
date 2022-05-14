package main

import (
	"testing"
)

func TestFlow(t *testing.T) {
	init4Tests(t)
	var flows FlowStruc
	snd := make([]FlowStepStruc, 1)
	snd[0] = FlowStepStruc{
		Act:   FlowActSnd,
		Name:  "echo",
		Block: true,
		Dest:  FlowVarSign + "receiver" + FlowVarSep + "dest" + FlowVarSign,
		Data:  FlowVarSign + "receiver" + FlowVarSep + "data" + FlowVarSign,
	}
	rcv := make([]FlowStepStruc, 1)
	rcv[0] = FlowStepStruc{
		Act:   FlowActRcv,
		Name:  "receiver",
		Block: true,
		Steps: snd,
	}
	flows.Conns = make([]FlowConnStruc, 2)
	flows.Conns[0] = FlowConnStruc{
		Protocol: *prot,
		Name:     "consumer",
		Steps:    rcv,
	}
	rcv1 := make([]FlowStepStruc, 1)
	rcv1[0] = FlowStepStruc{
		Act:   FlowActRcv,
		Block: true,
		Data:  FlowVarSign + FlowVarFil + FlowVarSep + "result.data" + FlowVarSign,
	}
	snd1 := make([]FlowStepStruc, 1)
	snd1[0] = FlowStepStruc{
		Act:   FlowActSnd,
		Block: true,
		Dest:  FlowVarSign + "peer" + FlowVarSign,
		Data:  "Bonjour",
		Name:  "sender",
		Steps: rcv1,
	}
	flows.Conns[1] = FlowConnStruc{
		Protocol: *prot,
		Name:     "producer",
		Peer:     FlowVarSign + "consumer" + FlowVarSep + "listen" + FlowVarSign,
		Steps:    snd1,
	}
	//t.Log(flows)
	runFlow(flows)
}
