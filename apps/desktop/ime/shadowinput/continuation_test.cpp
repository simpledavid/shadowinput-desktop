#include "continuation.h"
#include <stdexcept>
#include <string>
#include <iostream>
static void require(bool ok,const char *reason) { if(!ok)throw std::runtime_error(reason); }
int main() {
    ShadowContinuation context{};
    shadow_context_append(&context,"明天",1);
    shadow_context_append(&context,"我们",2);
    require(std::string(shadow_context_text(&context,3))=="明天我们","word context lost");
    for (int i=0;i<200;++i) shadow_context_append(&context,"中",3);
    require(shadow_runes(context.text)==128 && std::strlen(context.text)==384,"UTF-8 context not bounded");
    require(!*shadow_context_text(&context,30003),"idle context retained");
    shadow_context_append(&context,"今天",30004);shadow_context_clear(&context);
    require(!*context.text,"explicit cancellation retained text");
    shadow_context_append(&context,"你好\n",30005);
    require(!*context.text,"multiline text accepted");
    std::cout<<"PASS: committed-word context, UTF-8 bounds, expiry and reset\n";
}
