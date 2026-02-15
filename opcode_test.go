package jpath_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kode4food/jpath"
)

func TestOpcodeString(t *testing.T) {
	tests := []struct {
		op   jpath.Opcode
		want string
	}{
		{op: jpath.OpSegmentStart, want: "seg/start arg=endpc"},
		{op: jpath.OpSelectName, want: "sel/name"},
		{op: jpath.OpSelectSliceB11PN, want: "sel/slice/b11pn step- s+ e-"},
		{op: jpath.Opcode(255), want: "Opcode(255)"},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.want, tc.op.String())
	}
}
