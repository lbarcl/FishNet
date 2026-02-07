#pragma once
#if defined (_WIN32)
    #define _WINSOCK_DEPRECATED_NO_WARNINGS
    #include <winsock2.h>
    #include <ws2tcpip.h>

    #pragma comment(lib, "ws2_32.lib")

    using socket_t = SOCKET;

    inline bool net_init() {
        WSADATA wsa;
        return WSAStartup(MAKEWORD(2, 2), &wsa) == 0;
    }

    inline void net_cleanup() {
        WSACleanup();
    }

    inline int net_close(socket_t s) {
        return closesocket(s);
    }

    inline bool set_nonblocking(socket_t s) {
        u_long mode = 1;
        return ioctlsocket(s, FIONBIO, &mode);
    }

    #define NET_ERR WSAGetLastError()

#else

    #include <sys/socket.h>
    #include <arpa/inet.h>
    #include <unistd.h>
    #include <fcntl.h>

    using socket_t = int;

    inline bool net_init() {
        return true;
    }

    inline void net_cleanup() {

    }

    inline int net_close(socket_t s) {
        return close(s)
    }

    inline bool set_nonblocking(socket_t s) {
        int flags = fcntl(s, F_GETFL, 0);
        return fcntl(s, F_SETFL, flags | O_NONBLOCK) == 0;
    }

    #define NET_ERR errno

#endif
