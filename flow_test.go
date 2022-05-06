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
		Data: FlowVarSign + "incoming" + FlowVarSign,
	}
	rcv := make([]FlowStepStruc, 1)
	rcv[0] = FlowStepStruc{
		Act:   FlowActRcv,
		Loop:  0,
		Steps: snd,
	}
	flows.Conns = make([]FlowConnStruc, 2)
	flows.Conns[0] = FlowConnStruc{
		Protocol: *prot,
		Steps:    rcv,
	}
	snd1 := make([]FlowStepStruc, 1)
	snd1[0] = FlowStepStruc{
		Act:  FlowActSnd,
		Dest: FlowVarSign + "svr" + FlowVarSep + "local" + FlowVarSign,
		Data: "Bonjour",
	}
	flows.Conns[1] = FlowConnStruc{
		Protocol: *prot,
		Wait:     FlowVarSign + "svr" + FlowVarSep + "listen" + FlowVarSign,
		Steps:    snd1,
	}
	t.Log(flows)
	runFlow(flows)
}
