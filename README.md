# zoom-go

Port of https://github.com/benbalter/zoom-launcher to zoom. Go easy, it's my first Go project.

## Installation

To install, run:

```bash
$ go get github.com/benbalter/zoom-go/cmd/zoom
```

This will install a `zoom` executable file into `$GOPATH/bin/zoom`.

## Usage

Ensure `$GOPATH/bin` is in your `$PATH`, and run `zoom`! That's all.

## Authorization

The first time you run `zoom`, you will see instructions for how to create a Google app in the Developer Console, authorize it to access your calendar, download credentials, then import the credentials into `zoom`. After you import, you should be walked through the process of authorizing in the browser. Paste the authorization code back into your terminal, and vòila, `zoom` will be all configured for your next run.