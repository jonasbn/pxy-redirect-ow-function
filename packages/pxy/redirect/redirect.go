package main

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sirupsen/logrus"
)

type ResponseHeaders struct {
	Location    string `json:"location,omitempty"`
	ContentType string `json:"content-type,omitempty"`
}

type Response struct {
	StatusCode int             `json:"statusCode,omitempty"`
	Headers    ResponseHeaders `json:"headers,omitempty"`
	Body       string          `json:"body,omitempty"`
}

var logger = logrus.New()

// Input validation constants
const (
	maxInputLength    = 100
	maxFragmentLength = 50
)

// validFragmentPattern matches alphanumeric characters, hyphens, and underscores only
var validFragmentPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// validateInput performs basic input validation and sanitization
func validateInput(input string, maxLength int) error {
	if len(input) == 0 {
		return fmt.Errorf("input cannot be empty")
	}

	if len(input) > maxLength {
		return fmt.Errorf("input exceeds maximum length of %d characters", maxLength)
	}

	if !utf8.ValidString(input) {
		return fmt.Errorf("input contains invalid UTF-8 characters")
	}

	return nil
}

// validateFragment validates the fragment part of the URL
func validateFragment(fragment string) error {
	if err := validateInput(fragment, maxFragmentLength); err != nil {
		return err
	}

	if !validFragmentPattern.MatchString(fragment) {
		return fmt.Errorf("fragment contains invalid characters - only alphanumeric, hyphens, and underscores are allowed")
	}

	return nil
}

// sanitizeForHTML escapes HTML special characters to prevent XSS
func sanitizeForHTML(input string) string {
	return html.EscapeString(input)
}

// createSafeErrorMessage creates an error message with properly escaped user input
func createSafeErrorMessage(scheme, host, version, fragment string, messageType string) error {
	// Sanitize all user-controlled inputs
	safeScheme := sanitizeForHTML(scheme)
	safeHost := sanitizeForHTML(host)
	safeVersion := sanitizeForHTML(version)
	safeFragment := sanitizeForHTML(fragment)

	switch messageType {
	case "invalid_version":
		return fmt.Errorf(`<p>You only made it this far, because the specified URL requires a version number as the first part to redirect to the documentation</p><p>%s://%s/<span class="my-times">%s</span>/%s</p><p>In order to get the redirect to work, please specify both a version and a fragment</p><p>Example: <a href="https://pxy.fi/13/wall">https://pxy.fi/13/wall</a></p><p>See more information at: <a href="https://github.com/jonasbn/pxy-redirect-ow-function">GitHub</a></p>`,
			safeScheme, safeHost, safeVersion, safeFragment)
	case "insufficient_parts":
		return fmt.Errorf(`<p>You only made it this far, because the specified URL has insufficient parts to redirect to the documentation</p><p>%s://%s/<span class="my-times">%s</span></p><p>In order to get the redirect to work, please specify both a version and a fragment</p><p>Example: <a href="https://pxy.fi/13/wall">https://pxy.fi/13/wall</a></p><p>See more information at: <a href="https://github.com/jonasbn/pxy-redirect-ow-function">GitHub</a></p>`,
			safeScheme, safeHost, safeVersion)
	case "invalid_fragment":
		return fmt.Errorf(`<p>You only made it this far, because the specified URL requires a valid fragment as the second part to redirect to the documentation</p><p>%s://%s/%s/<span class="my-times">%s</span></p><p>In order to get the redirect to work, please specify both a version and a fragment</p><p>Example: <a href="https://pxy.fi/13/wall">https://pxy.fi/13/wall</a></p><p>See more information at: <a href="https://github.com/jonasbn/pxy-redirect-ow-function">GitHub</a></p>`,
			safeScheme, safeHost, safeVersion, safeFragment)
	default:
		return fmt.Errorf("Invalid URL format. Please check the documentation for proper usage.")
	}
}

// Can be uncommented for local testing
/* func main() {
	args := make(map[string]interface{})
	headers := make(map[string]interface{})

	// https://pxy.fi/6/wc++98-c++11-compat-binary-literal
	headers["user-agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko)"
	headers["do-connecting-ip"] = "192.168.1.2"
	headers["referer"] = "https://pxy.fi/6/wc++98-c++11-compat-binary-literal"
	headers["x-request-id"] = "4d84db433a35256e7fdd395f430a9121"

	args["__ow_path"] = "/6/wc++98-c++11-compat-binary-literal"
	args["__ow_headers"] = headers

	resp := Main(args)

	fmt.Printf("Response: %#v\n", resp)
} */

// Main is the entry point for the OpenWhisk action
func Main(args map[string]interface{}) Response {

	if os.Getenv("LOG_LEVEL") != "" {
		if os.Getenv("LOG_LEVEL") == "debug" {
			logger.SetLevel(logrus.DebugLevel)
		}
	}

	userAgent := ""
	ip := ""
	referer := ""
	path := ""
	requestID := ""

	if args["__ow_path"] != nil {
		path = args["__ow_path"].(string)
	}

	if args["__ow_headers"] != nil {
		requestHeaders := args["__ow_headers"]
		val, _ := requestHeaders.(map[string]interface{})

		if val["user-agent"] != nil {
			userAgent = val["user-agent"].(string)
		}

		if val["do-connecting-ip"] != nil {
			ip = val["do-connecting-ip"].(string)
		}

		if val["referer"] != nil {
			referer = val["referer"].(string)
		}

		if val["x-request-id"] != nil {
			requestID = val["x-request-id"].(string)
		}
	}

	logger.WithFields(logrus.Fields{
		"ip":         ip,
		"user-agent": userAgent,
		"referer":    referer,
		"request-id": requestID,
	}).Debugf("Running with log level: %s", logger.GetLevel())

	logger.WithFields(logrus.Fields{
		"ip":         ip,
		"user-agent": userAgent,
		"referer":    referer,
		"request-id": requestID,
	}).Infof("Received URL: >%s< via >%s<", path, referer)

	emitHeartbeat()

	url, err := parseRedirectURL(path)

	if err != nil {
		return Response{
			StatusCode: http.StatusInternalServerError, // 500
			Headers: ResponseHeaders{
				ContentType: "text/plain; charset=utf-8",
			},
			Body: err.Error(),
		}
	}

	targetURL, err := assembleTargetURL(url)

	if err != nil {
		return Response{
			StatusCode: http.StatusBadRequest, // 400
			Headers: ResponseHeaders{
				ContentType: "text/html; charset=utf-8",
			},
			Body: err.Error(),
		}
	}

	logger.WithFields(logrus.Fields{
		"ip":         ip,
		"user-agent": userAgent,
		"referer":    referer,
		"request-id": requestID,
	}).Infof("Redirecting to: >%s<", targetURL)

	return Response{
		Headers: ResponseHeaders{
			Location:    targetURL,
			ContentType: "text/plain; charset=utf-8",
		},
		StatusCode: http.StatusPermanentRedirect, // 308
		Body:       "redirecting...",
	}
}

func parseRedirectURL(path string) (*url.URL, error) {

	redirectURL, parseErr := url.Parse(path)

	if parseErr != nil {
		logger.Errorf("Unable to parse received URL: >%s<", path)
		return nil, fmt.Errorf("Unable to parse received URL: >%s<", path)
	}

	logger.Debugf("Parsed URL: >%s<", redirectURL)

	return redirectURL, nil
}

func assembleTargetURL(url *url.URL) (string, error) {

	s := strings.SplitN(url.Path, "/", 3)

	logger.Debugf("Parsed following parts: >%#v<", s)

	// 0 is empty because we split on "/" and the URL begins with "/"
	// 1 == version
	// 2 == fragment

	url.Host = "pxy.fi"
	url.Scheme = "https"

	// Check if we have insufficient parts before proceeding
	if len(s) < 3 {
		logger.Errorf("insufficient parts in provided url: >%s<", url.String())
		// Use safe fragment placeholder when not available
		fragmentPlaceholder := ""
		if len(s) >= 2 {
			return "", createSafeErrorMessage(url.Scheme, url.Host, s[1], fragmentPlaceholder, "insufficient_parts")
		}
		return "", fmt.Errorf("Invalid URL format. Please check the documentation for proper usage.")
	}

	majorlevel := s[1]
	fragment := s[2]

	// Validate the version part (majorlevel)
	if err := validateInput(majorlevel, maxInputLength); err != nil {
		logger.Errorf("invalid version input: %v", err)
		return "", createSafeErrorMessage(url.Scheme, url.Host, majorlevel, fragment, "invalid_version")
	}

	// Rewrite fragment and replacement of +
	// The command line flag:
	// -Wc++98-c++11-compat-binary-literal¶
	// has the anchor:
	// wc-98-c-11-compat-binary-literal
	fragment = strings.ReplaceAll(fragment, "+", "-")
	fragment = strings.ReplaceAll(fragment, "--", "-")

	// Validate the fragment part
	if err := validateFragment(fragment); err != nil {
		logger.Errorf("invalid fragment input: %v", err)
		return "", createSafeErrorMessage(url.Scheme, url.Host, majorlevel, fragment, "invalid_fragment")
	}

	// default patchlevel and minorlevel (see special treatment below)
	patchlevel := "0"
	minorlevel := "0"

	// Convert majorlevel to number to validate it is a number
	// And we need in for some additional logic, since the releases differ and we need to do some meddling
	// for versions 17 and up
	major, err := strconv.Atoi(majorlevel)
	if err != nil {
		logger.Errorf("first part of url: >%s< is not a number: %q", url.String(), s[1])
		return "", createSafeErrorMessage(url.Scheme, url.Host, majorlevel, fragment, "invalid_version")
	}

	// HACK: 17.0.0 was replaced with 17.0.1
	// So we have to link to: https://releases.llvm.org/17.0.1/tools/clang/docs/DiagnosticsReference.html
	// REF: https://github.com/llvm/llvm-project/releases/tag/llvmorg-17.0.1
	if major == 17 {
		patchlevel = "1"
	}

	// HACK: 18.0.0 was released as 18.1.0
	// We have to link to: https://releases.llvm.org/18.1.0/tools/clang/docs/DiagnosticsReference.html
	// REF: https://github.com/llvm/llvm-project/releases/tag/llvmorg-18.1.0
	// They have started making documentation for minor releases
	// This pattern continues for additional releases: 19, 20 and 21
	if major >= 18 {
		minorlevel = "1"
	}

	targetURL := fmt.Sprintf("https://releases.llvm.org/%s.%s.%s/tools/clang/docs/DiagnosticsReference.html#%s", majorlevel, minorlevel, patchlevel, fragment)

	logger.Debug("Constructed the following redirect target URL: ", targetURL)

	// Construct the final URL with validated inputs
	// The fragment is already validated by validateFragment function
	return targetURL, nil
}

func emitHeartbeat() {

	heartbeatToken := os.Getenv("HEARTBEAT_TOKEN")
	heartbeatTarget := os.Getenv("HEARTBEAT_TARGET")
	heartbeatTargetTimeoutStr := os.Getenv("HEARTBEAT_TARGET_TIMEOUT")
	heartbeatTargetTimeout := 10 // default timeout in seconds

	url := fmt.Sprintf("%s%s", heartbeatTarget, heartbeatToken)

	logger.Debug("Emitting heartbeat to URL: ", url)

	if heartbeatTargetTimeoutStr != "" {
		if val, err := strconv.Atoi(heartbeatTargetTimeoutStr); err == nil {
			heartbeatTargetTimeout = val
		} else {
			logger.Warnf("Invalid HEARTBEAT_TARGET_TIMEOUT value: %s, using default %d seconds", heartbeatTargetTimeoutStr, heartbeatTargetTimeout)
		}
	}

	if heartbeatToken == "" {
		logger.Debug("No heartbeat token configured")
		return
	}

	// Create HTTP client with timeout to prevent resource exhaustion
	client := &http.Client{
		Timeout: time.Duration(heartbeatTargetTimeout) * time.Second,
	}

	resp, err := client.Get(url)

	if err != nil {
		logger.Errorf("Unable to emit heartbeat to URL: %s, %s", url, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logger.Errorf("Emitting heartbeat failed: %s", resp.Status)
	}
}
