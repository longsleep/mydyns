Mydyns
==========

Mydyns implements a HTTP API to update a dynamic DNS zone by dynamically
adding or removing A and AAAA records from a zone.

## Build requirements

  - [Go](http://golang.org) >= 1.1.0

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

## DNS configuration and key

Mydyns sends updates to a upstream Bind DNS server using `nsupdate` utility to
send Dynamic DNS Update requests to a name server. This needs authentication
and that means you need to generate a dnssec key which is used to connect to
the DNS server and allows the update.

```bash
$ dnssec-keygen -a HMAC-MD5 -b 512 -n HOST your.dns.zone
```

This creates a public and private key. Add the public key to to allow updates
to your DNS zone, and use the private key file when starting `mydynsd`.

## Tokens

Mydyns usess token to authenticate update requests for hosts. The token
contains the user created the token and the host the token should update. To
ensure that the token cannot be modified, it is digitaly signed. The signing
key is passed as file on `mydynsd` startup via the `--secret` parameter.

## Startup

```bash
$ ./mydynsd \
	--server=your.name.server \
	--key=dnssec.key.private \
	--zone=your.dns.zone \
	--users=users.db \
	--hosts=hosts.db \
	--secret=secret.dat \
	--listen=127.0.0.1:8040 \
	--ttl=60
```

## HTTP API

The server provides HTTP API endpoints.

### Token

First one is /token which is used to generate a update token for a host. The
/token end point requires HTTP Basic authentication which provides the user
and password. When successfull the token is returned.

```bash
$ curl -u user:password https://yourserver/token?hostname=myhost
```

### Update

To send an update request use /update end point with the `token` parameter.
When no further parameters are passed, it will set the IP address where the
request came from for the hostname encoded in the token. You can also pass
the IP address manually with the `myip` parameter. For compatibility reasons
the `myip=auto` parameter is also supported. To only return the current IP
without changing anything, pass the `check` parameter.

```bash
$ curl https://yourserver/update?token=tokenvalue
```

There is a update script example in the `scripts` directory which you can
use to run from cron or similar.

--
Simon Eisenmann - mailto:simon@longsleep.org
