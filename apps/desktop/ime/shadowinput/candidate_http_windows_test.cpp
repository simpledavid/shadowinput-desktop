#include <winsock2.h>
#include <windows.h>
#include <future>
#include <iostream>
#include <stdexcept>
#include <string>

static std::wstring wide(const std::string &text) {
    return std::wstring(text.begin(),text.end());
}
#include "candidate_http_windows.h"

static void require(bool ok,const char *message) { if(!ok)throw std::runtime_error(message); }
struct Listener {
    SOCKET socket=::socket(AF_INET,SOCK_STREAM,IPPROTO_TCP);
    INTERNET_PORT port=0;
    Listener() {
        sockaddr_in address{};address.sin_family=AF_INET;address.sin_addr.s_addr=htonl(INADDR_LOOPBACK);
        require(socket!=INVALID_SOCKET && bind(socket,reinterpret_cast<sockaddr*>(&address),sizeof(address))==0 && listen(socket,1)==0,"loopback server failed");
        int length=sizeof(address);
        require(getsockname(socket,reinterpret_cast<sockaddr*>(&address),&length)==0,"server port missing");
        port=ntohs(address.sin_port);
    }
    ~Listener(){closesocket(socket);}
};

static void canceledRequestClosesConnection() {
    Listener listener;
    std::promise<void> accepted;
    auto acceptedFuture=accepted.get_future();
    auto server=std::async(std::launch::async,[&] {
        SOCKET peer=accept(listener.socket,nullptr,nullptr);
        if(peer==INVALID_SOCKET)return false;
        DWORD timeout=5000;
        setsockopt(peer,SOL_SOCKET,SO_RCVTIMEO,reinterpret_cast<const char*>(&timeout),sizeof(timeout));
        char buffer[4096];
        bool received=recv(peer,buffer,sizeof(buffer),0)>0;
        accepted.set_value();
        int count;
        do { count=recv(peer,buffer,sizeof(buffer),0); } while(count>0);
        bool closed=count==0 || WSAGetLastError()==WSAECONNRESET;
        closesocket(peer);
        return received && closed;
    });
    auto request=std::make_shared<ShadowCandidateRequest>(1,"synthetic phrase");
    auto client=std::async(std::launch::async,[&] {
        ShadowCandidateHTTP transport(request);
        return transport.run(listener.port,L"TESTONLYABCDEFGHIJKLMNOP");
    });
    require(acceptedFuture.wait_for(std::chrono::seconds(3))==std::future_status::ready,"request did not arrive");
    request->canceled=true;
    require(client.wait_for(std::chrono::seconds(2))==std::future_status::ready,"canceled request still waiting for inference");
    require(client.get().empty(),"canceled request produced output");
    require(server.get(),"cancellation did not close the HTTP connection");
}
static void successfulRequestReturnsBody() {
    Listener listener;
    auto server=std::async(std::launch::async,[&] {
        SOCKET peer=accept(listener.socket,nullptr,nullptr);
        if(peer==INVALID_SOCKET)return;
        char buffer[4096];recv(peer,buffer,sizeof(buffer),0);
        std::string response="HTTP/1.1 200 OK\r\nContent-Length: 15\r\nConnection: close\r\n\r\nfresh candidate";
        send(peer,response.data(),static_cast<int>(response.size()),0);
        shutdown(peer,SD_SEND);closesocket(peer);
    });
    auto request=std::make_shared<ShadowCandidateRequest>(2,"synthetic phrase");
    ShadowCandidateHTTP transport(request);
    require(transport.run(listener.port,L"TESTONLYABCDEFGHIJKLMNOP")==L"fresh candidate","normal response failed");
    server.get();
}
int main() {
    WSADATA data{};
    if(WSAStartup(MAKEWORD(2,2),&data)!=0)return 1;
    try {
        canceledRequestClosesConnection();
        successfulRequestReturnsBody();
        std::cout<<"PASS: real loopback WinHTTP cancellation and subsequent response\n";
    }catch(const std::exception &error){std::cerr<<error.what()<<'\n';WSACleanup();return 1;}
    WSACleanup();return 0;
}
