#include <jni.h>
#include <stdlib.h>
#include <string.h>

struct go_string {
    const char *str;
    long n;
};

extern char *hCashCreateAccount(struct go_string pwd);

extern char *hCashBurnProof();

JNIEXPORT jstring JNICALL
Java_com_hpb_android_backend_GoHCashBackend_hCashCreateAccount(JNIEnv *env,
                                                        jclass c, jstring pwd) {
    jstring ret;
    const char *pwd_str = (*env)->GetStringUTFChars(env, pwd, 0);
    size_t path_len = (*env)->GetStringUTFLength(env, pwd);
    char *accountRet = hCashCreateAccount((struct go_string) {
            .str = pwd_str,
            .n = path_len
    });
    (*env)->ReleaseStringUTFChars(env, pwd, pwd_str);

    ret = (*env)->NewStringUTF(env, accountRet);
    free(accountRet);
    return ret;
}

//export hCashBurnProof
JNIEXPORT jstring
Java_com_hpb_android_backend_GoHCashBackend_hCashBurnProof(JNIEnv *env, jclass c) {
    jstring ret;
    char *gRet = hCashBurnProof();
    if (!gRet)
        return NULL;
    ret = (*env)->NewStringUTF(env, gRet);
    free(gRet);
    return ret;
}
