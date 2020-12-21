#include <jni.h>
#include <stdlib.h>
#include <string.h>

struct go_string {
    const char *str;
    long n;
};

extern char *hCashCreateAccount(struct go_string secret);

extern char *hCashSign(struct go_string input);

extern int   hCashReadBalance(struct go_string param);

extern char *hCashShuffle(struct go_string input);

extern char *hCashTransferProof(struct go_string input);

extern char *hCashBurnProof(struct go_string input);


JNIEXPORT jstring JNICALL
Java_com_hpb_android_backend_GoHCashBackend_hCashCreateAccount(JNIEnv *env,
                                                        jclass c, jstring secret) {
    jstring ret;
    const char *secret_str = (*env)->GetStringUTFChars(env, secret, 0);
    size_t secret_len = (*env)->GetStringUTFLength(env, secret);
    char *accountRet = hCashCreateAccount((struct go_string) {
            .str = secret_str,
            .n = secret_len
    });
    (*env)->ReleaseStringUTFChars(env, secret, secret_str);

    ret = (*env)->NewStringUTF(env, accountRet);
    free(accountRet);
    return ret;
}

//extern char *hCashSign(struct go_string input);
JNIEXPORT jstring JNICALL
Java_com_hpb_android_backend_GoHCashBackend_hCashSign(JNIEnv *env,
                                                        jclass c, jstring input) {
    jstring ret;
    const char *input_str = (*env)->GetStringUTFChars(env, input, 0);
    size_t input_len = (*env)->GetStringUTFLength(env, input);
    char *signRet = hCashSign((struct go_string) {
            .str = input_str,
            .n = input_len
    });
    (*env)->ReleaseStringUTFChars(env, input, input_str);

    ret = (*env)->NewStringUTF(env, signRet);
    free(signRet);
    return ret;
}
//extern int   hCashReadBalance(struct go_string param);
JNIEXPORT jint JNICALL
Java_com_hpb_android_backend_GoHCashBackend_hCashReadBalance(JNIEnv *env,
                                                        jclass c, jstring input) {
    const char *input_str = (*env)->GetStringUTFChars(env, input, 0);
    size_t input_len = (*env)->GetStringUTFLength(env, input);
    int balance = hCashReadBalance((struct go_string) {
            .str = input_str,
            .n = input_len
    });
    (*env)->ReleaseStringUTFChars(env, input, input_str);

    return balance;
}

//extern char *hCashShuffle(struct go_string input);
JNIEXPORT jstring JNICALL
Java_com_hpb_android_backend_GoHCashBackend_hCashShuffle(JNIEnv *env,
                                                        jclass c, jstring input) {
    jstring ret;
    const char *input_str = (*env)->GetStringUTFChars(env, input, 0);
    size_t input_len = (*env)->GetStringUTFLength(env, input);
    char *shuffled = hCashShuffle((struct go_string) {
            .str = input_str,
            .n = input_len
    });
    (*env)->ReleaseStringUTFChars(env, input, input_str);

    ret = (*env)->NewStringUTF(env, shuffled);
    free(shuffled);
    return ret;
}

//extern char *hCashTransferProof(struct go_string input);
JNIEXPORT jstring JNICALL
Java_com_hpb_android_backend_GoHCashBackend_hCashTransferProof(JNIEnv *env,
                                                        jclass c, jstring input) {
    jstring ret;
    const char *input_str = (*env)->GetStringUTFChars(env, input, 0);
    size_t input_len = (*env)->GetStringUTFLength(env, input);
    char *proof = hCashTransferProof((struct go_string) {
            .str = input_str,
            .n = input_len
    });
    (*env)->ReleaseStringUTFChars(env, input, input_str);

    ret = (*env)->NewStringUTF(env, proof);
    free(proof);
    return ret;
}

//extern char *hCashBurnProof(struct go_string input);
JNIEXPORT jstring JNICALL
Java_com_hpb_android_backend_GoHCashBackend_hCashBurnProof(JNIEnv *env,
                                                        jclass c, jstring input) {
    jstring ret;
    const char *input_str = (*env)->GetStringUTFChars(env, input, 0);
    size_t input_len = (*env)->GetStringUTFLength(env, input);
    char *proof = hCashBurnProof((struct go_string) {
            .str = input_str,
            .n = input_len
    });
    (*env)->ReleaseStringUTFChars(env, input, input_str);

    ret = (*env)->NewStringUTF(env, proof);
    free(proof);
    return ret;
}