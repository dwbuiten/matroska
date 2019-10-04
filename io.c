/*
 * This file contains all the I/O callbacks that MatroskaParser requires
 * to operate. In practice these are thing wrappers around basic CRT
 * functions, or callbacks into Go-land, with a little glue.
 *
 * Also contains some simple allocation helps for CGO, and a function to
 * register the callbacks, since this is not possible from CGO.
 */

#include <stdlib.h>
#include <string.h>

#include "io.h"
#include "_cgo_export.h"

static int ioread(struct IO *cc, ulonglong pos, void *buffer, int count)
{
    if (count == 0)
        return 0;

    if (pos != cc->pos) {
        int ret = cSeekCallback(&(cc->key)[0], pos);
        if (ret < 0)
            return -1;
        cc->pos = pos;
    }

    int ret = cReadCallback(&(cc->key)[0], buffer, count);
    if (ret < 0)
        return -1;
    cc->pos += (ulonglong) count;

    return ret;
}

static longlong scan(struct IO *cc, ulonglong start, unsigned signature)
{
    return -1;
}

static unsigned getcachesize(struct IO *cc)
{
    return 64 * 1024;
}

static const char *geterror(struct IO *cc)
{
    return "muh error";
}

static void *memalloc(struct IO *cc, size_t size)
{
    return malloc(size);
}

static void *memrealloc(struct IO *cc, void *mem, size_t newsize)
{
    return realloc(mem, newsize);
}

static void memfree(struct IO *cc, void *mem)
{
    free(mem);
}

static int progress(struct IO *cc, ulonglong cur, ulonglong max)
{
    return 1;
}

static longlong getfilesize(struct IO *cc) {
    return cSizeCallback(&(cc->key)[0]);
}

void io_set_callbacks(IO *input, char *key)
{
    input->pos = 0;
    strncpy(&(input->key)[0], key, 36);

    input->input.read = (int (*)(InputStream *,ulonglong,void *,int))ioread;
    input->input.scan = (longlong (*)(InputStream *,ulonglong,unsigned int))scan;
    input->input.getcachesize = (unsigned (*)(InputStream *cc))getcachesize;
    input->input.geterror = (const char *(*)(InputStream *))geterror;
    input->input.memalloc = (void *(*)(InputStream *,size_t))memalloc;
    input->input.memrealloc = (void *(*)(InputStream *,void *,size_t))memrealloc;
    input->input.memfree = (void (*)(InputStream *,void *))memfree;
    input->input.progress = (int (*)(InputStream *,ulonglong,ulonglong))progress;
    input->input.getfilesize = (longlong (*)(InputStream *))getfilesize;
}

IO *io_alloc(void)
{
    return calloc(1, sizeof(IO));
}

void io_free(IO *io)
{
    free(io);
}
