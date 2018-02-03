; Copyright 2009 The Go Authors. All rights reserved.
; Use of this source code is governed by a BSD-style
; license that can be found in the LICENSE file.
PUBLIC crosscall_386

_TEXT SEGMENT
; void crosscall_386(void (*fn)(void))
;
; Calling into the 8c tool chain, where all registers are caller save.
; Called from standard x86 ABI, where %ebp, %ebx, %esi,
; and %edi are callee-save, so they must be saved explicitly.
crosscall_386 PROC
	pushl ebp
  movl esp, ebp
  pushl ebx
  pushl esi
  pushl edi

  movl ebp, eax	; fn
  call eax

  popl edi
  popl esi
  popl ebx
  popl ebp
	push epb
	ret

crosscall_386 ENDP
_TEXT ENDS

END
