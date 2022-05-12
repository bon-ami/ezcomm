package main

import (
	"testing"
)

func TestFlow(t *testing.T) {
	init4Tests(t)
	var flows FlowStruc
	snd := make([]FlowStepStruc, 1)
	snd[0] = FlowStepStruc{
		Act:  FlowActSnd,
		Dest: FlowVarSign + "peer" + FlowVarSign,
		Data: FlowVarSign + "receiver" + FlowVarSign + "data" + FlowVarSign,
	}
	rcv := make([]FlowStepStruc, 1)
	rcv[0] = FlowStepStruc{
		Act:   FlowActRcv,
		Loop:  0,
		Name:  "receiver",
		Steps: snd,
	}
	flows.Conns = make([]FlowConnStruc, 2)
	flows.Conns[0] = FlowConnStruc{
		Protocol: *prot,
		Name:     "consumer",
		Steps:    rcv,
	}
	snd1 := make([]FlowStepStruc, 1)
	snd1[0] = FlowStepStruc{
		Act:  FlowActSnd,
		Dest: FlowVarSign + "peer" + FlowVarSign,
		Data: "Bonjour",
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
