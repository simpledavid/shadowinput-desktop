#pragma once
#include <winhttp.h>
#include "candidate_state.h"

// Only the worker calls WinHTTP. UI cancellation sets the request's atomic flag;
// the worker closes an asynchronous request within one short wait interval.
class ShadowCandidateHTTP {
    HINTERNET session=nullptr,connection=nullptr,request=nullptr;
    HANDLE completed=CreateEventW(nullptr,FALSE,FALSE,nullptr);
    HANDLE unloaded=CreateEventW(nullptr,TRUE,FALSE,nullptr);
    bool notifyUnload=false;
    std::atomic<DWORD> status{0},received{0};
    std::shared_ptr<ShadowCandidateRequest> operation;
    ULONGLONG deadline=GetTickCount64()+16000;
    std::string payload;
    char buffer[1025]{};

    static void CALLBACK callback(HINTERNET,DWORD_PTR context,DWORD status,void *,DWORD count) {
        auto self=reinterpret_cast<ShadowCandidateHTTP*>(context);
        if(!self)return;
        self->received=count;
        self->status=status;
        SetEvent(self->completed);
    }
    bool wait(DWORD expected) {
        while(!operation->canceled && GetTickCount64()<deadline) {
            DWORD ready=WaitForSingleObject(completed,25);
            if(ready==WAIT_OBJECT_0)return !operation->canceled && status==expected;
            if(ready!=WAIT_TIMEOUT)return false;
        }
        return false;
    }
    bool open(INTERNET_PORT port) {
        if(!completed || !unloaded || operation->canceled)return false;
        session=WinHttpOpen(L"shadowinput IME",WINHTTP_ACCESS_TYPE_NO_PROXY,WINHTTP_NO_PROXY_NAME,WINHTTP_NO_PROXY_BYPASS,WINHTTP_FLAG_ASYNC);
        if(!session)return false;
        notifyUnload=WinHttpSetOption(session,WINHTTP_OPTION_UNLOAD_NOTIFY_EVENT,&unloaded,sizeof(unloaded))!=FALSE;
        if(!notifyUnload || !WinHttpSetTimeouts(session,1000,1000,1000,16000))return false;
        connection=WinHttpConnect(session,L"127.0.0.1",port,0);
        if(!connection)return false;
        request=WinHttpOpenRequest(connection,L"POST",L"/candidate",nullptr,WINHTTP_NO_REFERER,WINHTTP_DEFAULT_ACCEPT_TYPES,0);
        if(!request)return false;
        DWORD disabled=WINHTTP_DISABLE_REDIRECTS|WINHTTP_DISABLE_COOKIES|WINHTTP_DISABLE_AUTHENTICATION;
        if(!WinHttpSetOption(request,WINHTTP_OPTION_DISABLE_FEATURE,&disabled,sizeof(disabled)))return false;
        DWORD flags=WINHTTP_CALLBACK_FLAG_SENDREQUEST_COMPLETE|WINHTTP_CALLBACK_FLAG_HEADERS_AVAILABLE|WINHTTP_CALLBACK_FLAG_READ_COMPLETE|WINHTTP_CALLBACK_FLAG_REQUEST_ERROR;
        return WinHttpSetStatusCallback(request,callback,flags,0)!=WINHTTP_INVALID_STATUS_CALLBACK;
    }
public:
    explicit ShadowCandidateHTTP(std::shared_ptr<ShadowCandidateRequest> operation):operation(operation),payload(operation->source) {}
    ShadowCandidateHTTP(const ShadowCandidateHTTP&)=delete;
    ShadowCandidateHTTP &operator=(const ShadowCandidateHTTP&)=delete;
    ~ShadowCandidateHTTP() {
        if(request)WinHttpCloseHandle(request);
        if(connection)WinHttpCloseHandle(connection);
        if(session)WinHttpCloseHandle(session);
        // WinHTTP signals this only after its last callback returns. Keep both
        // callback context/buffers and the worker's DLL reference alive until then.
        if(notifyUnload)WaitForSingleObject(unloaded,INFINITE);
        if(unloaded)CloseHandle(unloaded);
        if(completed)CloseHandle(completed);
    }
    std::wstring run(INTERNET_PORT port,const std::wstring &token) {
        if(!open(port) || operation->canceled)return {};
        std::wstring headers=L"Content-Type: text/plain; charset=utf-8\r\nAuthorization: Bearer "+token+L"\r\n";
        DWORD size=static_cast<DWORD>(payload.size());
        if(!WinHttpSendRequest(request,headers.c_str(),static_cast<DWORD>(headers.size()),payload.data(),size,size,reinterpret_cast<DWORD_PTR>(this)) || !wait(WINHTTP_CALLBACK_STATUS_SENDREQUEST_COMPLETE))return {};
        if(!WinHttpReceiveResponse(request,nullptr) || !wait(WINHTTP_CALLBACK_STATUS_HEADERS_AVAILABLE))return {};
        DWORD code=0,length=sizeof(code);
        if(!WinHttpQueryHeaders(request,WINHTTP_QUERY_STATUS_CODE|WINHTTP_QUERY_FLAG_NUMBER,nullptr,&code,&length,nullptr) || code!=200)return {};
        std::string body;
        for(;;) {
            if(operation->canceled || !WinHttpReadData(request,buffer,sizeof(buffer),nullptr) || !wait(WINHTTP_CALLBACK_STATUS_READ_COMPLETE))return {};
            DWORD count=received.load();
            if(!count)break;
            if(count>sizeof(buffer) || body.size()+count>1024)return {};
            body.append(buffer,count);
        }
        return body.find_first_of("\r\n\0",0,3)==std::string::npos?wide(body):std::wstring{};
    }
};
