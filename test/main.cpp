#include "tcpserver.hpp"

int main() {
    TCPServer* server = new TCPServer("0.0.0.0", 8080, 100, 1);
    while (true) {
        server->Accept();
    }
}
