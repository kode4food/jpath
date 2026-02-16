package jpath

// Opcode identifies a VM operation
type Opcode uint8

const (
	// Slice tag decode
	// F/B = step sign (>0 / <0)
	// XY  = hasStart/hasEnd (0/1)
	// P/N = explicit bound sign (>=0 / <0)

	OpSegmentStart     Opcode = iota // seg/start arg=endpc
	OpSegmentEnd                     // seg/end
	OpDescend                        // descend/self+deep
	OpSelectName                     // sel/name
	OpSelectIndex                    // sel/index
	OpSelectWildcard                 // sel/wildcard
	OpSelectArrayAll                 // sel/array/all
	OpSelectSliceF00                 // sel/slice/f00 step+ s0 e0
	OpSelectSliceF10P                // sel/slice/f10p step+ s+ e0
	OpSelectSliceF10N                // sel/slice/f10n step+ s- e0
	OpSelectSliceF01P                // sel/slice/f01p step+ s0 e+
	OpSelectSliceF01N                // sel/slice/f01n step+ s0 e-
	OpSelectSliceF11PP               // sel/slice/f11pp step+ s+ e+
	OpSelectSliceF11PN               // sel/slice/f11pn step+ s+ e-
	OpSelectSliceF11NP               // sel/slice/f11np step+ s- e+
	OpSelectSliceF11NN               // sel/slice/f11nn step+ s- e-
	OpSelectSliceB00                 // sel/slice/b00 step- s0 e0
	OpSelectSliceB10P                // sel/slice/b10p step- s+ e0
	OpSelectSliceB10N                // sel/slice/b10n step- s- e0
	OpSelectSliceB01P                // sel/slice/b01p step- s0 e+
	OpSelectSliceB01N                // sel/slice/b01n step- s0 e-
	OpSelectSliceB11PP               // sel/slice/b11pp step- s+ e+
	OpSelectSliceB11PN               // sel/slice/b11pn step- s+ e-
	OpSelectSliceB11NP               // sel/slice/b11np step- s- e+
	OpSelectSliceB11NN               // sel/slice/b11nn step- s- e-
	OpSelectSliceEmpty               // sel/slice/empty step=0
	OpSelectFilter                   // sel/filter
)
