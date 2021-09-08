// Code generated by command: go run gen.go -sse. DO NOT EDIT.

#include "textflag.h"

DATA prime_sse<>+0(SB)/4, $0x9e3779b1
DATA prime_sse<>+4(SB)/4, $0x9e3779b1
DATA prime_sse<>+8(SB)/4, $0x9e3779b1
DATA prime_sse<>+12(SB)/4, $0x9e3779b1
GLOBL prime_sse<>(SB), RODATA|NOPTR, $16

// func accumSSE(acc *[8]uint64, data *byte, key *byte, len uint64)
// Requires: SSE2
TEXT ·accumSSE(SB), NOSPLIT, $0-32
	MOVQ  acc+0(FP), AX
	MOVQ  data+8(FP), CX
	MOVQ  key+16(FP), DX
	MOVQ  key+16(FP), BX
	MOVQ  len+24(FP), SI
	MOVOU (AX), X1
	MOVOU 16(AX), X2
	MOVOU 32(AX), X3
	MOVOU 48(AX), X4
	MOVOU prime_sse<>+0(SB), X0

accum_large:
	CMPQ    SI, $0x00000400
	JLE     accum
	MOVOU   (CX), X5
	MOVOU   (DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   16(CX), X5
	MOVOU   16(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   32(CX), X5
	MOVOU   32(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   48(CX), X5
	MOVOU   48(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   64(CX), X5
	MOVOU   8(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   80(CX), X5
	MOVOU   24(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   96(CX), X5
	MOVOU   40(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   112(CX), X5
	MOVOU   56(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   128(CX), X5
	MOVOU   16(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   144(CX), X5
	MOVOU   32(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   160(CX), X5
	MOVOU   48(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   176(CX), X5
	MOVOU   64(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   192(CX), X5
	MOVOU   24(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   208(CX), X5
	MOVOU   40(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   224(CX), X5
	MOVOU   56(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   240(CX), X5
	MOVOU   72(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   256(CX), X5
	MOVOU   32(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   272(CX), X5
	MOVOU   48(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   288(CX), X5
	MOVOU   64(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   304(CX), X5
	MOVOU   80(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   320(CX), X5
	MOVOU   40(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   336(CX), X5
	MOVOU   56(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   352(CX), X5
	MOVOU   72(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   368(CX), X5
	MOVOU   88(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   384(CX), X5
	MOVOU   48(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   400(CX), X5
	MOVOU   64(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   416(CX), X5
	MOVOU   80(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   432(CX), X5
	MOVOU   96(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   448(CX), X5
	MOVOU   56(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   464(CX), X5
	MOVOU   72(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   480(CX), X5
	MOVOU   88(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   496(CX), X5
	MOVOU   104(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   512(CX), X5
	MOVOU   64(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   528(CX), X5
	MOVOU   80(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   544(CX), X5
	MOVOU   96(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   560(CX), X5
	MOVOU   112(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   576(CX), X5
	MOVOU   72(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   592(CX), X5
	MOVOU   88(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   608(CX), X5
	MOVOU   104(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   624(CX), X5
	MOVOU   120(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   640(CX), X5
	MOVOU   80(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   656(CX), X5
	MOVOU   96(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   672(CX), X5
	MOVOU   112(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   688(CX), X5
	MOVOU   128(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   704(CX), X5
	MOVOU   88(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   720(CX), X5
	MOVOU   104(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   736(CX), X5
	MOVOU   120(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   752(CX), X5
	MOVOU   136(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   768(CX), X5
	MOVOU   96(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   784(CX), X5
	MOVOU   112(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   800(CX), X5
	MOVOU   128(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   816(CX), X5
	MOVOU   144(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   832(CX), X5
	MOVOU   104(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   848(CX), X5
	MOVOU   120(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   864(CX), X5
	MOVOU   136(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   880(CX), X5
	MOVOU   152(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   896(CX), X5
	MOVOU   112(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   912(CX), X5
	MOVOU   128(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   928(CX), X5
	MOVOU   144(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   944(CX), X5
	MOVOU   160(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   960(CX), X5
	MOVOU   120(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   976(CX), X5
	MOVOU   136(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   992(CX), X5
	MOVOU   152(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   1008(CX), X5
	MOVOU   168(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	ADDQ    $0x00000400, CX
	SUBQ    $0x00000400, SI
	MOVOU   X1, X5
	PSRLQ   $0x2f, X5
	PXOR    X5, X1
	MOVOU   128(DX), X5
	PXOR    X5, X1
	PSHUFD  $0xf5, X1, X5
	PMULULQ X0, X1
	PMULULQ X0, X5
	PSLLQ   $0x20, X5
	PADDQ   X5, X1
	MOVOU   X2, X5
	PSRLQ   $0x2f, X5
	PXOR    X5, X2
	MOVOU   144(DX), X5
	PXOR    X5, X2
	PSHUFD  $0xf5, X2, X5
	PMULULQ X0, X2
	PMULULQ X0, X5
	PSLLQ   $0x20, X5
	PADDQ   X5, X2
	MOVOU   X3, X5
	PSRLQ   $0x2f, X5
	PXOR    X5, X3
	MOVOU   160(DX), X5
	PXOR    X5, X3
	PSHUFD  $0xf5, X3, X5
	PMULULQ X0, X3
	PMULULQ X0, X5
	PSLLQ   $0x20, X5
	PADDQ   X5, X3
	MOVOU   X4, X5
	PSRLQ   $0x2f, X5
	PXOR    X5, X4
	MOVOU   176(DX), X5
	PXOR    X5, X4
	PSHUFD  $0xf5, X4, X5
	PMULULQ X0, X4
	PMULULQ X0, X5
	PSLLQ   $0x20, X5
	PADDQ   X5, X4
	JMP     accum_large

accum:
	CMPQ    SI, $0x40
	JLE     finalize
	MOVOU   (CX), X0
	MOVOU   (BX), X5
	PXOR    X0, X5
	PSHUFD  $0x31, X5, X6
	PMULULQ X5, X6
	PSHUFD  $0x4e, X0, X0
	PADDQ   X0, X1
	PADDQ   X6, X1
	MOVOU   16(CX), X0
	MOVOU   16(BX), X5
	PXOR    X0, X5
	PSHUFD  $0x31, X5, X6
	PMULULQ X5, X6
	PSHUFD  $0x4e, X0, X0
	PADDQ   X0, X2
	PADDQ   X6, X2
	MOVOU   32(CX), X0
	MOVOU   32(BX), X5
	PXOR    X0, X5
	PSHUFD  $0x31, X5, X6
	PMULULQ X5, X6
	PSHUFD  $0x4e, X0, X0
	PADDQ   X0, X3
	PADDQ   X6, X3
	MOVOU   48(CX), X0
	MOVOU   48(BX), X5
	PXOR    X0, X5
	PSHUFD  $0x31, X5, X6
	PMULULQ X5, X6
	PSHUFD  $0x4e, X0, X0
	PADDQ   X0, X4
	PADDQ   X6, X4
	ADDQ    $0x00000040, CX
	SUBQ    $0x00000040, SI
	ADDQ    $0x00000008, BX
	JMP     accum

finalize:
	CMPQ    SI, $0x00
	JE      return
	SUBQ    $0x40, CX
	ADDQ    SI, CX
	MOVOU   (CX), X0
	MOVOU   121(DX), X5
	PXOR    X0, X5
	PSHUFD  $0x31, X5, X6
	PMULULQ X5, X6
	PSHUFD  $0x4e, X0, X0
	PADDQ   X0, X1
	PADDQ   X6, X1
	MOVOU   16(CX), X0
	MOVOU   137(DX), X5
	PXOR    X0, X5
	PSHUFD  $0x31, X5, X6
	PMULULQ X5, X6
	PSHUFD  $0x4e, X0, X0
	PADDQ   X0, X2
	PADDQ   X6, X2
	MOVOU   32(CX), X0
	MOVOU   153(DX), X5
	PXOR    X0, X5
	PSHUFD  $0x31, X5, X6
	PMULULQ X5, X6
	PSHUFD  $0x4e, X0, X0
	PADDQ   X0, X3
	PADDQ   X6, X3
	MOVOU   48(CX), X0
	MOVOU   169(DX), X5
	PXOR    X0, X5
	PSHUFD  $0x31, X5, X6
	PMULULQ X5, X6
	PSHUFD  $0x4e, X0, X0
	PADDQ   X0, X4
	PADDQ   X6, X4

return:
	MOVOU X1, (AX)
	MOVOU X2, 16(AX)
	MOVOU X3, 32(AX)
	MOVOU X4, 48(AX)
	RET

// func accumBlockSSE(acc *[8]uint64, data *byte, key *byte)
// Requires: SSE2
TEXT ·accumBlockSSE(SB), NOSPLIT, $0-24
	MOVQ    acc+0(FP), AX
	MOVQ    data+8(FP), CX
	MOVQ    key+16(FP), DX
	MOVOU   (AX), X1
	MOVOU   16(AX), X2
	MOVOU   32(AX), X3
	MOVOU   48(AX), X4
	MOVOU   prime_sse<>+0(SB), X0
	MOVOU   (CX), X5
	MOVOU   (DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   16(CX), X5
	MOVOU   16(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   32(CX), X5
	MOVOU   32(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   48(CX), X5
	MOVOU   48(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   64(CX), X5
	MOVOU   8(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   80(CX), X5
	MOVOU   24(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   96(CX), X5
	MOVOU   40(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   112(CX), X5
	MOVOU   56(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   128(CX), X5
	MOVOU   16(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   144(CX), X5
	MOVOU   32(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   160(CX), X5
	MOVOU   48(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   176(CX), X5
	MOVOU   64(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   192(CX), X5
	MOVOU   24(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   208(CX), X5
	MOVOU   40(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   224(CX), X5
	MOVOU   56(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   240(CX), X5
	MOVOU   72(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   256(CX), X5
	MOVOU   32(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   272(CX), X5
	MOVOU   48(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   288(CX), X5
	MOVOU   64(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   304(CX), X5
	MOVOU   80(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   320(CX), X5
	MOVOU   40(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   336(CX), X5
	MOVOU   56(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   352(CX), X5
	MOVOU   72(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   368(CX), X5
	MOVOU   88(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   384(CX), X5
	MOVOU   48(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   400(CX), X5
	MOVOU   64(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   416(CX), X5
	MOVOU   80(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   432(CX), X5
	MOVOU   96(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   448(CX), X5
	MOVOU   56(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   464(CX), X5
	MOVOU   72(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   480(CX), X5
	MOVOU   88(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   496(CX), X5
	MOVOU   104(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   512(CX), X5
	MOVOU   64(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   528(CX), X5
	MOVOU   80(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   544(CX), X5
	MOVOU   96(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   560(CX), X5
	MOVOU   112(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   576(CX), X5
	MOVOU   72(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   592(CX), X5
	MOVOU   88(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   608(CX), X5
	MOVOU   104(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   624(CX), X5
	MOVOU   120(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   640(CX), X5
	MOVOU   80(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   656(CX), X5
	MOVOU   96(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   672(CX), X5
	MOVOU   112(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   688(CX), X5
	MOVOU   128(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   704(CX), X5
	MOVOU   88(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   720(CX), X5
	MOVOU   104(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   736(CX), X5
	MOVOU   120(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   752(CX), X5
	MOVOU   136(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   768(CX), X5
	MOVOU   96(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   784(CX), X5
	MOVOU   112(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   800(CX), X5
	MOVOU   128(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   816(CX), X5
	MOVOU   144(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   832(CX), X5
	MOVOU   104(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   848(CX), X5
	MOVOU   120(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   864(CX), X5
	MOVOU   136(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   880(CX), X5
	MOVOU   152(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   896(CX), X5
	MOVOU   112(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   912(CX), X5
	MOVOU   128(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   928(CX), X5
	MOVOU   144(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   944(CX), X5
	MOVOU   160(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   960(CX), X5
	MOVOU   120(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X1
	PADDQ   X7, X1
	MOVOU   976(CX), X5
	MOVOU   136(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X2
	PADDQ   X7, X2
	MOVOU   992(CX), X5
	MOVOU   152(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X3
	PADDQ   X7, X3
	MOVOU   1008(CX), X5
	MOVOU   168(DX), X6
	PXOR    X5, X6
	PSHUFD  $0x31, X6, X7
	PMULULQ X6, X7
	PSHUFD  $0x4e, X5, X5
	PADDQ   X5, X4
	PADDQ   X7, X4
	MOVOU   X1, X5
	PSRLQ   $0x2f, X5
	PXOR    X5, X1
	MOVOU   128(DX), X5
	PXOR    X5, X1
	PSHUFD  $0xf5, X1, X5
	PMULULQ X0, X1
	PMULULQ X0, X5
	PSLLQ   $0x20, X5
	PADDQ   X5, X1
	MOVOU   X2, X5
	PSRLQ   $0x2f, X5
	PXOR    X5, X2
	MOVOU   144(DX), X5
	PXOR    X5, X2
	PSHUFD  $0xf5, X2, X5
	PMULULQ X0, X2
	PMULULQ X0, X5
	PSLLQ   $0x20, X5
	PADDQ   X5, X2
	MOVOU   X3, X5
	PSRLQ   $0x2f, X5
	PXOR    X5, X3
	MOVOU   160(DX), X5
	PXOR    X5, X3
	PSHUFD  $0xf5, X3, X5
	PMULULQ X0, X3
	PMULULQ X0, X5
	PSLLQ   $0x20, X5
	PADDQ   X5, X3
	MOVOU   X4, X5
	PSRLQ   $0x2f, X5
	PXOR    X5, X4
	MOVOU   176(DX), X5
	PXOR    X5, X4
	PSHUFD  $0xf5, X4, X5
	PMULULQ X0, X4
	PMULULQ X0, X5
	PSLLQ   $0x20, X5
	PADDQ   X5, X4
	MOVOU   X1, (AX)
	MOVOU   X2, 16(AX)
	MOVOU   X3, 32(AX)
	MOVOU   X4, 48(AX)
	RET
