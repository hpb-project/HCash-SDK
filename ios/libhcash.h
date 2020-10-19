#ifndef HCASH_H
#define HCASH_H

#include <sys/types.h>
#include <stdint.h>
#include <stdbool.h>

typedef struct { const char *p; size_t n; } gostring_t;
//typedef void(*logger_fn_t)(int level, const char *msg);
extern char* hCashBurnProof();
extern char* hCashCreateAccount(gostring_t pwd);
#endif
