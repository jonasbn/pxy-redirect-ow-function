# pxy-redirect-ow-function

This is an experimental serverless implementation of the [pxy-redirect service][SERVICE].

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
https://pxy.fk/p/r/5/rsanitize-address
```

Redirects to:

```text
https://releases.llvm.org/5.0.0/tools/clang/docs/DiagnosticsReference.html#rsanitize-address
```

Do note the version number is expanded from a single digit to a 3 part version number.

This scheme is used by [clang diagnostic flags matrix generator][GENERATOR], but by the service, since I have not found a way to shorted the function URL.

Please see my [blog post][BLOG] for the long version.

## Resources and References

[GENERATOR]: https://github.com/jonasbn/clang-diagnostic-flags-matrix
[SERVICE]: https://github.com/jonasbn/pxy-redirect
[BLOG]: https://dev.to/jonasbn/challenges-solutions-and-more-challenges-and-more-solutions-4j3f
