#pragma once
#include <stdint.h>
#include <string.h>

enum { ShadowDebounceMS=180, ShadowContextLifetimeMS=30000, ShadowContextRunes=128, ShadowSuffixRunes=24 };

// Only this input method's committed text, never surrounding document contents.
// The platform owns the caret anchor; this value owns bounded, expiring UTF-8.
typedef struct { char text[ShadowContextRunes*4+1]; uint64_t time; } ShadowContinuation;
static inline size_t shadow_runes(const char *text) {
    size_t count=0;
    for (;*text;++text) if (((unsigned char)*text&0xc0)!=0x80) ++count;
    return count;
}
static inline void shadow_context_clear(ShadowContinuation *context) { memset(context,0,sizeof(*context)); }
static inline const char *shadow_context_text(ShadowContinuation *context,uint64_t now) {
    if (now<context->time || now-context->time>=ShadowContextLifetimeMS) shadow_context_clear(context);
    return context->text;
}
static inline void shadow_context_append(ShadowContinuation *context,const char *text,uint64_t now) {
    size_t length=strlen(text);
    if (!length || length>=sizeof(context->text)) { shadow_context_clear(context); return; }
    for (size_t i=0;i<length;++i) if ((unsigned char)text[i]<32 || text[i]==127) { shadow_context_clear(context); return; }
    const char *previous=shadow_context_text(context,now);
    char joined[sizeof(context->text)*2];
    size_t before=strlen(previous);
    memcpy(joined,previous,before); memcpy(joined+before,text,length+1);
    size_t skip=0,count=shadow_runes(joined);
    while (count>ShadowContextRunes) {
        ++skip;
        while (((unsigned char)joined[skip]&0xc0)==0x80) ++skip;
        --count;
    }
    memcpy(context->text,joined+skip,before+length-skip+1); context->time=now;
}
