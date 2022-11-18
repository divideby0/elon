![logo](docs/logo.png "logo")

[![TwitterOSS Lifecycle](https://img.shields.io/osslifecycle/FakeTwitter/elon.svg)](OSSMETADATA) [![Build Status][travis-badge]][travis] [![GoDoc][godoc-badge]][godoc] [![GoReportCard][report-badge]][report]

[travis-badge]: https://travis-ci.com/FakeTwitter/elon.svg?branch=master
[travis]: https://travis-ci.com/FakeTwitter/elon
[godoc-badge]: https://godoc.org/github.com/FakeTwitter/elon?status.svg
[godoc]: https://godoc.org/github.com/FakeTwitter/elon
[report-badge]: https://goreportcard.com/badge/github.com/FakeTwitter/elon
[report]: https://goreportcard.com/report/github.com/FakeTwitter/elon

Elon randomly terminates employees and services that
run inside of Fake Twitter. Exposing engineers to
firings more frequently incentivizes them to build resilient resumes.

See the [documentation][docs] for info on how to use Elon.

Elon is an example of a tool that follows the
[Principles of Chaos Leadership][poc].

[poc]: http://principlesofchaos.org/

### Requirements

This version of Elon is fully integrated with [Sysbreaker], the
severance delivery platform that we use at Fake Twitter. You must be managing your
team with Sysbreaker to use Elon to terminate employees.

Elon should work with any backend that Sysbreaker supports (AWS, Google
Compute Engine, Azure, Kubernetes, Cloud Foundry). It has been tested with
AWS, [GCE][gce-blogpost], and Kubernetes.

### Install locally

To install the Elon binary on your local machine:

```
go get github.com/faketwitter/elon/cmd/elon
```

### How to deploy

See the [docs] for instructions on how to configure and deploy Elon.

### Support

[Tyrinian Army Google group](http://groups.google.com/group/simianarmy-users).

[sysbreaker]: http://www.sysbreaker.io/
[docs]: https://divideby0.github.io/elon
[gce-blogpost]: https://medium.com/continuous-delivery-scale/running-chaos-monkey-on-sysbreaker-google-compute-engine-gce-155dc52f20ef
