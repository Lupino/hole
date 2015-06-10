A hole to pass through the gateway.
==================================

When I visit raspberry pi' ssh server on some places,
I must set port forwarding on the home route, and set a dynamic DNS.
If the route is not your's, you will helpless.

I think it may have another way, so I try ssh port forwarding `ssh -CfNgR remote-port:localhost:local-port user@remote`, then vist the remote-port.

The hole is an other way similar ssh port forwarding.
On the global server create a `hole-server`, and the target host start `hole-local`.
Last you can visit the `hole-server` to replace the real server.

The hole suit the situation: A(private) can connect B(global), C(private) can connect B,
but B can't connect C, B can't connect A, and A can't connect C.

Install
-------

    go get -v github.com/Lupino/hole/cmd/hole-server
    go get -v github.com/Lupino/hole/cmd/hole-local

Quick start
-----------

    # Start on B
    hole-server -addr=tcp://B-IP:B-PORT
    # Start on C
    hole-local -addr=tcp://B-IP:B-PORT -src=tcp://localhost:C-PORT
    # On A just vist tcp://B-IP:B-PORT to replace visit C server

Example:
-------

    # A IP is 192.168.1.101
    # B IP is 120.26.120.168
    # C IP is 172.17.3.10

    # Now on B server
    hole-server -addr=tcp://120.26.120.168:4000

    # On C server
    hole-local -addr=tcp://120.26.120.168:4000 -src=tcp://127.0.0.1:22

    # On A server
    ssh root@120.26.120.168 -p 4000

    # Now A can visit C via B server
