#pragma once
#include "sock.h"

class TCPServer {
private:
  socket_t sockid;

public:
  TCPServer(const char *host, u_short port, int max_connection, int backlog);
};
