# How to contribute

Welcome to SuperEdge!

## Code of Conduct

Please do check our [Code of Conduct](./CODE_OF_CONDUCT.md) before making contributions.

# Email and chat

- Email: [superedge-email](https://groups.google.com/g/superedge)
- Slack: [superedge-community](https://join.slack.com/t/superedge-workspace/shared_invite/zt-k1kekpdz-jih6w8RByoylnfTmCTZpkA)

## Getting started

- Fork the repository on GitHub
- Read the README.md for build instructions

## Reporting bugs and creating issues

Reporting [bugs](https://github.com/superedge/superedge/issues) is one of the best ways to contribute.

## Contribution flow

This is a rough outline of what a contributor's workflow looks like:

- Create a topic branch from where to base the contribution. This is usually master.
- Make commits of logical units.
- Make sure commit messages are in the proper format (see below).
- Push changes in a topic branch to a personal fork of the repository.
- Submit a pull request to [superedge](https://github.com/superedge/superedge/pulls).
- The PR must receive a LGTM from two maintainers found in the MAINTAINERS file.

Thanks for contributing!

### Preparation

- ***FORK*** [SuperEdge](https://github.com/superedge/superedge) repository. The `Fork` button is in the top right corner of [superedge](https://github.com/superedge/superedge) home page.
- ***CLONE*** repository. `git clone https://github.com/<yourname>/superedge.git`
- ***SET REMOTE***.
```
git remote add upstream https://github.com/superedge/superedge.git
git remote set-url --push upstream no-pushing
```

### Code style

The coding style suggested by the Golang community is used in superedge. See the [style doc](https://github.com/golang/go/wiki/CodeReviewComments) for details.

### Format of the commit message

We follow a rough convention for commit messages that is designed to answer two
questions: what changed and why. The subject line should feature the what and
the body of the commit should describe the why.

```
tunnel: add grpc interceptor to log info on incoming requests

To improve debuggability of superedge. Added a grpc interceptor to log
info on incoming requests to tunnel server. The log output includes
remote client info, request content (with value field redacted), request
handling latency, response size, etc. Uses zap logger if available,
otherwise uses capnslog.

Fixes #10
```

The format can be described more formally as follows:

```
<package>: <what changed>
<BLANK LINE>
<why this change was made>
<BLANK LINE>
<footer>
```

### Pull request across multiple files and packages

If multiple files in a package are changed in a pull request for example:

```
tunnel/tunnel.go
edgeadm/edgeadm.go
```

At the end of the review process if multiple commits exist for a single package they
should be squashed/rebased into a single commit before being merged.

```
tunnel: <what changed>
[..]
```

If a pull request spans many packages these commits should be squashed/rebased into a single
commit using message with a more generic `*:` prefix.

```
*: <what changed>
[..]
```