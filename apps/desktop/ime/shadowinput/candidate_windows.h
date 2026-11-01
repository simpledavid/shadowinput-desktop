#pragma once
#include <winhttp.h>
#include <memory>
#include <mutex>
#include <fstream>
#include <atomic>
// Worker state never holds a TSF/COM pointer. A DLL reference protects each
// outstanding request; teardown revokes results without blocking the editor.
static bool shadowAIEnabled() {
 PWSTR base=nullptr;
 if(FAILED(SHGetKnownFolderPath(FOLDERID_RoamingAppData,0,nullptr,&base)))return false;
 std::wstring file=std::wstring(base)+L"\\shadowinput\\candidate.endpoint";CoTaskMemFree(base);
 return GetFileAttributesW(file.c_str())!=INVALID_FILE_ATTRIBUTES;
}
class ShadowCandidateClient {
 struct State { std::mutex mutex; bool alive=true,active=false; uint64_t revision=0,started=0; ULONGLONG changed=0; std::string source; std::wstring result; };
 std::shared_ptr<State> state=std::make_shared<State>();
 struct Work { std::shared_ptr<State> state; uint64_t revision; std::string source; HMODULE module; };
 static DWORD WINAPI work(void *raw) {
  std::unique_ptr<Work> job(static_cast<Work*>(raw));
  std::wstring result;
  PWSTR base=nullptr;
  if(SUCCEEDED(SHGetKnownFolderPath(FOLDERID_RoamingAppData,0,nullptr,&base))) {
   std::wstring file=std::wstring(base)+L"\\shadowinput\\candidate.endpoint";CoTaskMemFree(base);
   HANDLE input=CreateFileW(file.c_str(),GENERIC_READ,FILE_SHARE_READ|FILE_SHARE_WRITE|FILE_SHARE_DELETE,nullptr,OPEN_EXISTING,FILE_ATTRIBUTE_NORMAL,nullptr);
   char bytes[256]{};DWORD count=0;
   if(input!=INVALID_HANDLE_VALUE){if(!ReadFile(input,bytes,sizeof(bytes)-1,&count,nullptr))count=0;CloseHandle(input);}
   std::string config(bytes,count);auto end=config.find('\n'),next=config.find('\n',end==std::string::npos?0:end+1);
   if(end!=std::string::npos&&next!=std::string::npos&&config.rfind("127.0.0.1:",0)==0) {
    auto digits=config.substr(10,end-10),token=config.substr(end+1,next-end-1);
    unsigned long port=0;if(!digits.empty()&&digits.size()<=5&&digits.find_first_not_of("0123456789")==std::string::npos)port=std::strtoul(digits.c_str(),nullptr,10);
    if(port>0&&port<=65535&&token.size()>=20&&token.size()<100&&token.find_first_not_of("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ234567") == std::string::npos) {
     HINTERNET session=WinHttpOpen(L"shadowinput IME",WINHTTP_ACCESS_TYPE_NO_PROXY,WINHTTP_NO_PROXY_NAME,WINHTTP_NO_PROXY_BYPASS,0);
     if(session){WinHttpSetTimeouts(session,1000,1000,1000,16000);
      HINTERNET connection=WinHttpConnect(session,L"127.0.0.1",static_cast<INTERNET_PORT>(port),0);
      HINTERNET request=connection?WinHttpOpenRequest(connection,L"POST",L"/candidate",nullptr,WINHTTP_NO_REFERER,WINHTTP_DEFAULT_ACCEPT_TYPES,0):nullptr;
      if(request){DWORD disable=WINHTTP_DISABLE_REDIRECTS;WinHttpSetOption(request,WINHTTP_OPTION_DISABLE_FEATURE,&disable,sizeof(disable));
       std::wstring headers=L"Content-Type: text/plain; charset=utf-8\r\nAuthorization: Bearer "+wide(token)+L"\r\n";
       if(WinHttpSendRequest(request,headers.c_str(),static_cast<DWORD>(headers.size()),job->source.data(),static_cast<DWORD>(job->source.size()),static_cast<DWORD>(job->source.size()),0)&&WinHttpReceiveResponse(request,nullptr)) {
        DWORD status=0,size=sizeof(status);
        if(WinHttpQueryHeaders(request,WINHTTP_QUERY_STATUS_CODE|WINHTTP_QUERY_FLAG_NUMBER,nullptr,&status,&size,nullptr)&&status==200){
         std::string body;char buffer[1025];DWORD n=0;bool ok=true;
         while(ok&&(ok=WinHttpReadData(request,buffer,sizeof(buffer),&n))&&n){body.append(buffer,n);if(body.size()>1024){ok=false;break;}}
         if(ok&&body.find_first_of("\r\n\0",0,3)==std::string::npos)result=wide(body);
        }
       }WinHttpCloseHandle(request);
      }if(connection)WinHttpCloseHandle(connection);WinHttpCloseHandle(session);
     }
    }
   }
  }
  {std::lock_guard<std::mutex> lock(job->state->mutex);job->state->active=false;
   if(job->state->alive&&job->state->revision==job->revision&&result.size()<=256)job->state->result=result;}
  HMODULE module=job->module;job.reset();FreeLibraryAndExitThread(module,0);return 0;
 }
public:
 ~ShadowCandidateClient(){std::lock_guard<std::mutex> lock(state->mutex);state->alive=false;state->result.clear();}
 void update(const std::string &source,uint64_t revision){std::lock_guard<std::mutex> lock(state->mutex);state->source=source;state->revision=revision;state->changed=GetTickCount64();state->result.clear();}
 std::wstring poll(uint64_t revision){
  std::lock_guard<std::mutex> lock(state->mutex);
  if(state->revision!=revision)return {};
  if(!state->active&&state->started!=revision&&state->source.size()>=12&&state->source.size()<=512&&GetTickCount64()-state->changed>=700){
   HMODULE module=nullptr;
   if(GetModuleHandleExW(GET_MODULE_HANDLE_EX_FLAG_FROM_ADDRESS,reinterpret_cast<LPCWSTR>(&work),&module)){
    state->started=revision;state->active=true;
    auto job=new Work{state,revision,state->source,module};
    HANDLE thread=CreateThread(nullptr,0,work,job,0,nullptr);
    if(thread)CloseHandle(thread);else{delete job;state->active=false;FreeLibrary(module);}
   }
  }
  return state->result;
 }
};
