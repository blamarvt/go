; Copyright 2009 The Go Authors. All rights reserved.
; Use of this source code is governed by a BSD-style
; license that can be found in the LICENSE file.
PUBLIC crosscall_amd64

_TEXT SEGMENT
 ; void crosscall_amd64(void (*fn)(void))
 ;
 ; Calling into the 6c tool chain, where all registers are caller save.
 ; Called from standard x86-64 ABI, where %rbx, %rbp, %r12-%r15
 ; are callee-save so they must be saved explicitly.
 ; The standard x86-64 ABI passes the three arguments m, g, fn
 ; in %rdi, %rsi, %rdx.
crosscall_amd64 PROC
	push rbx
	push rbp
	push r12
	push r13
	push r14
	push r15

	call rcx ; fn

	pop r15
	pop r14
	pop r13
	pop r12
	pop rbp
	pop rbx
	ret

crosscall_amd64 ENDP
_TEXT ENDS

END
