# pxy-redirect-ow-function

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
https://pxy.fi/p/r/5/rsanitize-address
```

Redirects to:

```text
https://releases.llvm.org/5.0.0/tools/clang/docs/DiagnosticsReference.html#rsanitize-address
```

Do note the version number is expanded from a single digit to a 3 part version number.

This scheme is used by [clang diagnostic flags matrix generator][GENERATOR].

Please see my [blog post][BLOG] for the long version.

## Resources and References

- [DigitalOcean][DO]
- [clang diagnostic flags matrix generator][GENERATOR]
- [My TIL collection: clang diagnostic flags](https://github.com/jonasbn/til/blob/master/clang/diagnostic_flags.md) (GitHub)
- [My TIL collection: clang diagnostic flags](http://jonasbn.github.io/til/clang/diagnostic_flags.html) (website)
- [pxy-redirect][SERVICE]
- [pxy.fi][PXYFI]

[GENERATOR]: https://github.com/jonasbn/clang-diagnostic-flags-matrix
[SERVICE]: https://github.com/jonasbn/pxy-redirect
[BLOG]: https://dev.to/jonasbn/challenges-solutions-and-more-challenges-and-more-solutions-4j3f
[DO]: https://www.digitalocean.com/
[PXYFI]: https://pxy.fi/p/r/
