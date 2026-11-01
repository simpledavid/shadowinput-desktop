#include "candidate_state.h"
#include <iostream>
#include <stdexcept>

static void require(bool ok,const char *reason) { if(!ok)throw std::runtime_error(reason); }
int main() {
    try {
        ShadowCandidateState state;
        state.update("first phrase",1,0);
        require(!state.begin(1,699),"request started before debounce");
        auto first=state.begin(1,700);
        require(bool(first),"first request missing");
        state.update("second phrase",2,710);
        require(first->canceled,"editing did not cancel the old request");
        auto second=state.begin(2,1410);
        require(bool(second),"new phrase is blocked by the old request");
        state.complete(first,L"stale");
        require(state.read(2).empty(),"old result was delivered");
        require(!state.begin(2,2000),"old completion released the new request");
        state.complete(second,L"fresh");
        require(state.read(2)==L"fresh","new result missing");
        state.update("third phrase",3,2000);
        auto third=state.begin(3,2700);
        state.update({},3,2710);
        require(third->canceled,"clearing composition did not cancel transport");
        state.complete(third,L"late");
        require(state.read(3).empty(),"same-revision clear allowed stale delivery");
        state.update("fourth phrase",4,3000);
        auto fourth=state.begin(4,3700);
        state.close();
        require(fourth->canceled,"teardown did not cancel transport");
        state.complete(fourth,L"late");
        require(state.read(4).empty(),"teardown accepted a result");
        std::cout<<"PASS: debounce, replacement, cancellation, stale completion and teardown\n";
    } catch(const std::exception &error) {
        std::cerr<<error.what()<<'\n';return 1;
    }
}
