#pragma once
#include <atomic>
#include <cstdint>
#include <memory>
#include <mutex>
#include <string>

struct ShadowCandidateRequest {
    const uint64_t revision;
    const std::string source;
    std::atomic<bool> canceled{false};
    ShadowCandidateRequest(uint64_t revision,const std::string &source):revision(revision),source(source) {}
};

// Composition state is shared with workers, never with TSF/COM objects.
class ShadowCandidateState {
    std::mutex mutex;
    bool alive=true;
    uint64_t revision=0,started=0,changed=0;
    std::string source;
    std::wstring result;
    std::shared_ptr<ShadowCandidateRequest> current;
    void revoke() {
        if(current)current->canceled=true;
        current.reset();
    }
public:
    void update(const std::string &text,uint64_t version,uint64_t now) {
        std::lock_guard<std::mutex> lock(mutex);
        revoke();
        source=text;revision=version;changed=now;result.clear();
    }
    std::shared_ptr<ShadowCandidateRequest> begin(uint64_t version,uint64_t now) {
        std::lock_guard<std::mutex> lock(mutex);
        if(!alive || version!=revision || current || started==revision || source.size()<12 || source.size()>512 || now-changed<700)return {};
        started=revision;
        current=std::make_shared<ShadowCandidateRequest>(revision,source);
        return current;
    }
    void complete(const std::shared_ptr<ShadowCandidateRequest> &request,const std::wstring &text) {
        std::lock_guard<std::mutex> lock(mutex);
        if(current!=request || request->canceled)return;
        current.reset();
        if(alive && revision==request->revision && text.size()<=256)result=text;
    }
    std::wstring read(uint64_t version) {
        std::lock_guard<std::mutex> lock(mutex);
        return version==revision?result:std::wstring{};
    }
    void close() {
        std::lock_guard<std::mutex> lock(mutex);
        revoke();
        alive=false;result.clear();
    }
};
