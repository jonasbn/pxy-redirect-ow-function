# pxy-redirect-ow-function

## Introduction

This is an experimental serverless implementation of the [pxy-redirect service][SERVICE] I have created for deployment on [DigitalOcean][DO].

It's goal is to take a short URL following a required format and redirect to the designated URL.

The scheme is:

```text
<domain>/<version>/<fragment>
```

Redirects to:

```text
https://<domain>/<version>/tools/clang/docs/DiagnosticsReference.html#<fragment>
```

Example:

```text
https://pxy.fi/5/rsanitize-address
```

Redirects to:

```text
https://releases.llvm.org/5.0.0/tools/clang/docs/DiagnosticsReference.html#rsanitize-address
```

Do note the version number is expanded from a single digit to a 3 part version number.

This scheme is used by [clang diagnostic flags matrix generator][GENERATOR]. Please see my [blog post][BLOG] for the long version.

**pxy-redirect-ow-function** is transparent, so introduction of a version `16.0.0` would work out of the box. The pages utilizing the service and the generator however rely on human interaction in order to be updated.

- [clang diagnostic flags matrix generator][GENERATOR]
- [My TIL collection: clang diagnostic flags](http://jonasbn.github.io/til/clang/diagnostic_flags.html) (website)

## Diagnostics

This is a collection of errors which can be emitted from the service. Not all are visible to the end user and not all error scenarios are documented.

This section and documentation is primarily aimed and what can be recovered from.

### Unable to assemble URL (`400`)

This is the most common error it will provide additional information as to why the request was regarded as a bad request.

#### **insufficient parts in provided url**

The the request does not contain enough parts to assemble the redirect target URL.

The URL should consist of 2 parts.

1. Version number
2. Fragment

```text
https://pxy.fi/<version number>/<fragment>
```

Do note the version number is expanded from a single digit to a 3 part version number.

Luckily command line options (fragments) are only introduced or removed in major versions (X.0.0).

#### **first part of url is not a number**

The first part of the URL should be a number (integer), which is translated to a version number.

```text
https://pxy.fi/<version number>/<fragment>
```

Do note the version number is expanded from a single digit to a 3 part version number.

Not all numbers are supported since documentation for all versions is not available.

To my knowledge version ranging from `4.0.0` to `18.1.0` are supported, for reference these would be `4` and `18`.

Please visit the [releases.llvm.org][LLVM] website for more details and try the service, since the documentation might now be up to date with the service as for the supported versions.

#### **second part of url is not a string**

The second part of the URL should be a string.

```text
https://pxy.fi/<version number>/<fragment>
```

The second part is the fragment.

Please visit the [releases.llvm.org][LLVM] website for more details.

An example of a good fragment, which is available in all versions is: `wall`

```text
https://pxy.fi/4/wall
```

Redirects to:

```text
https://releases.llvm.org/4.0.0/tools/clang/docs/DiagnosticsReference.html#wall
```

The version number can be exchanged for a number between `4` and `15`.

## Logging

The default log level is `INFO`

It can be set to `DEBUG` via an environment variable, see below.

See the article on [log levels][LOGLEVELS] for more information.

The logging is currently collected on [logtail](https://betterstack.com/logtail), from [Better Stack](https://betterstack.com/).

## Monitoring

Currently the service is monitored using [Better Uptime](https://betteruptime.com/) from [Better Stack](https://betterstack.com/).

A public status page is [available](https://status.pxy.fi/).

It monitors the following:

- The reverse proxy (Nginx) via a health check
- Calling a redirectable URL and checking the response code. The URL goes via the reverse proxy (Nginx) and the service (OpenWhisk)
- The service (OpenWhisk) via a heartbeat

## Environment Variables

If the environment variable `LOG_LEVEL` is specified as `debug` the log level will be set to `DEBUG`.

## Resources and References

- [DigitalOcean][DO]
- [clang diagnostic flags matrix generator][GENERATOR]
- [My TIL collection: clang diagnostic flags](https://github.com/jonasbn/til/blob/master/clang/diagnostic_flags.md) (GitHub)
- [My TIL collection: clang diagnostic flags](http://jonasbn.github.io/til/clang/diagnostic_flags.html) (website)
- [pxy-redirect][SERVICE] service
- [pxy.fi][PXYFI] site
- [llvm releases documentation site][LLVM]
- [Log Levels][LOGLEVELS]

[GENERATOR]: https://github.com/jonasbn/clang-diagnostic-flags-matrix
[SERVICE]: https://github.com/jonasbn/pxy-redirect
[BLOG]: https://dev.to/jonasbn/challenges-solutions-and-more-challenges-and-more-solutions-4j3f
[DO]: https://www.digitalocean.com/
[PXYFI]: https://pxy.fi/p/r/
[LLVM]: https://releases.llvm.org/
[LOGLEVELS]: https://betterstack.com/community/guides/logging/log-levels-explained/
