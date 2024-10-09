linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o ./server_node .

windows:
	go build -o com.yessvpn.server-node.exe

#.PHONY: linux	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o ./server_node  .