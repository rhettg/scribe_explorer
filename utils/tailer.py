import json
import socket

if __name__ == "__main__":
	family, socktype, proto, canonname, sockaddr = socket.getaddrinfo("localhost", 3535, 0, socket.SOCK_STREAM, 0)[0]
	fd = socket.socket(family, socktype, proto)
	fd.connect(sockaddr)


	fd.send(json.dumps({"filters": [], "fields": ["unique_request_id", "uri"]}))
	fd.send("\n")

	while True:
		print fd.recv(1024),