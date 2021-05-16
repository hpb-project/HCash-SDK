#ifndef HCASH_H
#define HCASH_H

#include <sys/types.h>
#include <stdint.h>
#include <stdbool.h>

typedef struct { const char *p; size_t n; } gostring_t;
//typedef void(*logger_fn_t)(int level, const char *msg);
extern char *hCashCreateAccount(gostring_t secret);
extern char *hCashSign(gostring_t input);
extern int   hCashReadBalance(gostring_t param);
extern char *hCashShuffle(gostring_t input);
extern char *hCashTransferProof(gostring_t input);
extern char *hCashBurnProof(gostring_t input);

extern char *hCashTxRegister(gostring_t input);
extern char *hCashTxFund(gostring_t input);
extern char *hCashTxTransfer(gostring_t input);
extern char *hCashTxBurn(gostring_t input);
extern char *hCashTxSimulateAccounts(gostring_t input);

extern char *hCashParseSimulateAccountsData(gostring_t input);

#endif
