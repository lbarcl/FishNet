#include "tcpserver.hpp"
#include <iostream>

TCPServer::TCPServer(const char *host, u_short port, int max_connection, int backlog)
{
  if (!net_init()) {
    std::cerr << "WSAStartup failed\n";
    return;
  }

  sockid = socket(AF_INET, SOCK_STREAM, 0);
  if (sockid == -1) {
    std::cerr << "Socket creation failed\n";
    return;
  }

  set_nonblocking(sockid);

  sockaddr_in addr{};
  addr.sin_family = AF_INET;
  addr.sin_port = htons(port);

  if (strcmp(host, "0.0.0.0") == 0)
    addr.sin_addr.s_addr = INADDR_ANY;
  else if (inet_pton(AF_INET, host, &addr.sin_addr) <= 0) {
    std::cerr << "Invalid IP" << std::endl;
    return;
  }

  if (bind(sockid, (sockaddr *)&addr, sizeof(addr)) < 0) {
    std::cerr << "Bind failed:" << NET_ERR << std::endl;
    return;
  }

  if (listen(sockid, backlog) < 0) {
    std::cerr << "Listen failed:" << NET_ERR << std::endl;
    return;
  }

  std::cout << "TCP server listening on " << host << ":" << port << "\n";
}
