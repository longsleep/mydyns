Mydyns
==========

Mydyns implements a HTTP API to update a dynamic DNS zone by  adding or
removing A and AAAA records from a DNS zone. Mydyns uses the `nsupdate`
utility to submit Dynamic DNS Update requests as defined in RFC 2136 to a name
server.

## Build requirements

  - [Go](http://golang.org) >= 1.1.0


## Runtime requirements

  - nsupdate (Found in dnsutils provided with BIND)


## Building

```bash
$ make
```

## Configuration of users and hosts

Mydyns requires a users database and a hosts database. Both are simple text
files.

### Users database users.db

The users database can be managed with htpasswd from Apache. Make sure to use
SHA for password hashing.

```bash
$ htpasswd -c -s users.db myuser
```

### Hosts database hosts.db

The hosts database is a simple text file listing one host per line. In
addition the hosts have to be mapped to users. Users are added after the host
seperate by colon and comma separated.

```
somehost:usera,userb
otherhost:userc
```

### Security database security.db

The security database is a simple text file listing one user with the current
security code for this user. The entry is optional. The security code can
be used to expire all existing tokens for this user. Tokens must always have
the current security code, else they are not valid and useless. Change or set
the security code, if a token becomes stolen. All tokens of a single user use
the same security code.

```
usera:current security code
userb:supercode
```

## DNS configuration and key

Mydyns sends updates to a upstream Bind DNS server using `nsupdate` utility to
send Dynamic DNS Update requests to a name server. This needs authentication
and that means you need to generate a dnssec key which is used to connect to
the DNS server and allows the update.

```bash
$ dnssec-keygen -a HMAC-SHA256 -b 256 -n HOST your.dns.zone
```

This creates a public and private key. Add the public key to to allow updates
to your DNS zone, and use the private key file when starting `mydynsd`.


## Tokens

Mydyns usess tokens to authenticate /update requests for hosts. The token
contains a HMAC of the user and the host. The secret for creating the HMAC is
is read from a file passed via the `--secret` parameter. You shoud generate
the file with some random data.

```bash
$ dd if=/dev/urandom of=secret.key bs=1 count=32
```

The length of the key should be 32 or 64 bytes.


## Startup

```bash
$ ./mydynsd \
	--server=your.name.server \
	--key=dnssec.key.private \
	--zone=your.dns.zone \
	--users=users.db \
	--hosts=hosts.db \
	--security=security.db \
	--secret=secret.key \
	--listen=127.0.0.1:8040 \
	--ttl=60
```

While the server is running, you can send the HUP signal to make it reload the
database files for users, hosts and security. All other changes require a full
restart.

## HTTP API

The server provides HTTP API end points.

### /token

First one is /token which is used to generate a update token for a host. The
/token end point requires HTTP Basic authentication to provide the user and
password. When successfully authenticated and the user is listed in the hosts
database for the provides hostname, the token is returned. This tokenvalue can
then be used to use /update that hostname.

```bash
$ curl -u user:password https://yourserver/token?hostname=myhost
```

### /update

To send an update request use /update end point with the `token` parameter.
When no further parameters are passed, it will set the IP address where the
request came from for the hostname encoded in the token. You can also pass
the IP address manually with the `myip` parameter. For compatibility reasons
the value `auto` and the `address` parameter are also supported. To only
return the current IP without changing anything, pass the `check` parameter.

```bash
$ curl https://yourserver/update?token=tokenvalue
```

There is a update script example in the `scripts` directory which you can
use to run from cron or similar. Also check the `extra` directory for some
ideas to run the daemon as an upstart service.


## Expose service to the Internet

Mydyns runs on the local interface by default. If you want to expose the
service to the public internet you should run it behind a transparent proxy
like Nginx to provide TLS encryption. For auto detection of the remote IP
addresses to work, make sure that the proxy injects the remote IP address as
`X-Real-IP` HTTP request header.

### Nginx example

```
location ~* /(token|update)$ {
	proxy_pass http://127.0.0.1:38040;
	proxy_set_header Host $http_host;
	proxy_set_header X-Real-IP $remote_addr;
}
```

--
Simon Eisenmann - mailto:simon@longsleep.org
