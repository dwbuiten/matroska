#ifndef _IO_H
#define _IO_H

#include "MatroskaParser.h"

typedef struct IO {
    InputStream input;
    ulonglong pos;
    char key[37];
} IO;


void io_set_callbacks(IO *input, char *key);
IO *io_alloc(void);
void io_free(IO *io);

#endif

