tmate-bootstrap
===============

Source for tmate-bootstrap.cfapps.io

Location installation
---------------------

Copy `scripts/cf-ssh` to your local `$PATH`. Make it executable.

Usage
-----

The `cf-ssh` script is a helpful tool to create a dedicated Cloud Foundry container, with the target application's services and environment variables and SSH into it.

It creates a new Cloud Foundry application (appname-ssh), rather than trying to modify/reuse an existing production Cloud Foundry application.

From within your Cloud Foundry application project folder:

```
$ cf-ssh appname
Using manifest file ./cf-ssh.yml

Creating app appname-ssh

Uploading appname-ssh...

Binding service service1 to app appname-ssh
Binding service service2 to app appname-ssh

Starting app appname-ssh

Running: ssh GQtRhMTrxRhk0aYIJoqHVuBnz@ny.tmate.io
```

After you exit the SSH session, the `appname-ssh` application is deleted. This ensures that future SSH sessions will include up-to-date source code, buildpack, services and environment variables.

You can keep the container (albeit in "stopped" status) by setting the environment variable:

```
export CF_SSH_CLEANUP=keep
```

Note: currently it does not know how to specify a buildpack or discover the buildpack used by the target application.

Development
-----------

Install `go-bindata`:

```
go get github.com/jteeuwen/go-bindata/...
```

Generate `bindata.go` to include the `payload/payload.tgz` file:

```
go-bindata payload
```

Create `tmate-bootstrap` executable in current folder:

```
go build
```

Move the executable into the `http_server/payload` folder.

```
mv tmate-bootstrap http_server/payload/
```

Deploy to local Cloud Foundry
-----------------------------

There is a publicly available tmate-bootstrap server at https://tmate-bootstrap.cfapps.io

If you want to host it, and its `tmate-bootstrap` CLI payload, on your own Cloud Foundry:

```
cd http_server
bundle
cf push
```
