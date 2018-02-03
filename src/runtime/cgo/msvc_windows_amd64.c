// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#define WIN64_LEAN_AND_MEAN
#pragma warning(disable:4668 4255 4710) 
#include <windows.h>
#include <process.h>
#include <stdlib.h>
#include <stdio.h>
#pragma warning(default:4668 4255 4710) 
#include "libcgo.h"

static void threadentry(void*);

/* 2MB is default stack size for 64-bit Windows.
   Allocation granularity on Windows is typically 64 KB.
   The constant is also hardcoded in cmd/ld/pe.c (keep synchronized). */
#define STACKSIZE (2*1024*1024)

void
x_cgo_init(G *g)
{
	int tmp;
	g->stacklo = (uintptr)&tmp - STACKSIZE + 8*1024;
}


void
_cgo_sys_thread_start(ThreadStart *ts)
{
	uintptr_t thandle;

	thandle = _beginthread(threadentry, 0, ts);
	if(thandle == -1) {
		fprintf(stderr, "runtime: failed to create new OS thread (%d)\n", errno);
		abort();
	}
}

static void
threadentry(void *v)
{
	ThreadStart ts;

	ts = *(ThreadStart*)v;
	free(v);

	ts.g->stackhi = (uintptr)&ts;
	ts.g->stacklo = (uintptr)&ts - STACKSIZE + 8*1024;

	/*
	 * Set specific keys in thread local storage.
	 */
	__writegsqword(0x28, (unsigned long long) ts.tls);
    *(void **)ts.tls = (void *) ts.g;

	crosscall_amd64(ts.fn);
}
