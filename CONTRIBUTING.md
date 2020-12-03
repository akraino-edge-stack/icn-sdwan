# How to Contribute

SDEWAN is [Apache 2.0 licensed](LICENSE) and accepts contributions via ICN
gerrit repositry.[https://gerrit.akraino.org/r/admin/repos/icn/sdwan](
https://gerrit.akraino.org/r/admin/repos/icn/sdwan)                
                                                                               
The github page [https://github.com/akraino-edge-stack/icn-sdwan](https://github.com/akraino-edge-stack/icn-sdwan)
is github mirror repository for the user.                                      
                                                                               
Contributor login with Linux Foundation ID into [https://gerrit.akraino.org/r/admin/repos/icn/sdwan](https://gerrit.akraino.org/r/admin/repos/icn/sdwan)
and add their ssh key or use the LF username and password

Please follows the [Akraino's Gerrit Guide](https://wiki.akraino.org/display/AK/Gerrit+Guide)

This document outlines some of the conventions on development
workflow, commit message formatting, contact points and other resources to make
it easier to get your contribution accepted.

## Coding Style

Please follows the standard formatting recommendations and language idioms set out
in [Effective Go](https://golang.org/doc/effective_go.html) and in the
[Go Code Review Comments wiki](https://github.com/golang/go/wiki/CodeReviewComments).

## Certificate of Origin

In order to get a clear contribution chain of trust we use the [signed-off-by language](https://01.org/community/signed-process)
used by the Linux kernel project.

## Format of the patch

Beside the signed-off-by footer, we expect each patch to comply with the following format:

```
Change summary

More detailed explanation of your changes: Why and how.
Wrap it to 72 characters.
See [here] (http://chris.beams.io/posts/git-commit/)
for some more good advices.

Fixes #NUMBER (or URL to the issue)

Signed-off-by: <contributor@foo.com>
```

For example:

```
Fix poorly named identifiers
  
One identifier, fnname, in func.go was poorly named.  It has been renamed
to fnName.  Another identifier retval was not needed and has been removed
entirely.

Fixes #1
    
Signed-off-by: Abc Xyz <abc.xyz@intel.com>
```

## Pull requests

We accept patches, please follows the steps here - [Akraino's Gerrit Guide](https://wiki.akraino.org/display/AK/Gerrit+Guide)

## Email and Chat

The project uses the Slack chat:
- Slack: #[akraino-icn](https://akraino-icn-admin.herokuapp.com/) channel on slack
