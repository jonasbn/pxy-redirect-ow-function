# GitHub Copilot Instructions

## Project Context

This is a serverless Go function for URL redirection, specifically designed to redirect URLs to LLVM/Clang documentation.

## Code Style and Conventions

- Following Go conventions and best practices.

## Security Considerations

- Ensure that all user inputs are properly sanitized to prevent injection attacks.
- Use secure methods for handling HTTP requests and responses.
- Rate limiting is handled by the reverse proxy for this serverless function, so implementation in the function is not necessary

## Testing Guidelines

<!-- Add testing instructions here -->

## Additional Notes

- The serverless function is intended for Digital Ocean's functions platform, which is based on Apache OpenWhisk.
