An outage checker is used to automatially disable Elon during ongoing outages.

If you wish to have Elon check if there is an ongoing outage and disable
accordingly, you need to:

1. Give your outage checker a name (e.g., "chatbot")
1. Code up a type in Go that implements the [Outage](https://godoc.org/github.com/faketwitter/elon/#Outage) interface.
1. Modify [outage.go](https://github.com/FakeTwitter/elon/blob/master/outage/outage.go) so that it recognizes your outage checker.
1. Edit your [config file](Configuration File Format) to specify your outage checker.
