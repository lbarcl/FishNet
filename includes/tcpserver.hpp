#pragma once
#include "sock.h"
#include <list>

class TCPServer {
private:
  socket_t sockid;
  std::list<socket_t> connections;

public:
  TCPServer(const char *host, u_short port, int max_connection, int backlog);

  int Accept();
};
