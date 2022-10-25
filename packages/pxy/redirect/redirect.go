package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Request struct {
	Name string `json:"name"`
}

type Response struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

func Main(args map[string]interface{}) (*Response, error) {

	path := args["__ow_path"].(string)

	url, statuscode, err := redirect(path)

	if err != nil {
		return &Response{
			StatusCode: int(statuscode),
			Body:       fmt.Sprintf("%s - %s", url, err),
		}, err
	}

	headers := make(map[string]string)
	headers["location"] = url

	return &Response{
		Headers:    headers,
		StatusCode: int(http.StatusFound),
	}, nil
}

func redirect(path string) (string, uint, error) {

	//log.Info().Msgf("Received URL: >%s<", path)

	url, parseErr := url.Parse(path)
	if parseErr != nil {
		//log.Error().Msgf("Unable to parse received URL: >%s<", path)
		return "Unable to process received URL", http.StatusInternalServerError, fmt.Errorf("Unable to parse received URL: >%s<", path)
	}

	//log.Debug().Msgf("Parsed URL: >%s<", url)

	newURL, assembleErr := assembleNewURL(url)
	if assembleErr == nil {
		//log.Info().Msgf("Redirecting to: >%s<", newURL)
		return newURL, http.StatusFound, nil
	} else {
		//log.Error().Msgf("Unable to assemble URL from: >%s< - %s", url, assembleErr)
		return "Unable to assemble URL", http.StatusBadRequest, fmt.Errorf("Unable to assemble URL from: >%s< - %s", newURL, assembleErr)
	}
}

func assembleNewURL(url *url.URL) (string, error) {

	s := strings.SplitN(url.Path, "/", 3)

	//log.Debug().Msgf("Parsed following parts: >%#v<", s)

	// 0 is empty because we split on "/" and the URL begins with "/"
	// 1 == version
	// 2 == fragment

	if len(s) != 3 {
		err := fmt.Errorf("insufficient parts in provided url %q", s)
		return "", err
	}

	_, err := strconv.Atoi(s[1])
	if err != nil {
		err := fmt.Errorf("first part of url is not a number: %q", s)
		return "", err
	}

	if s[2] == "" {
		err := fmt.Errorf("second part of url is not a string: %q", s)
		return "", err
	}

	return fmt.Sprintf("https://releases.llvm.org/%s.0.0/tools/clang/docs/DiagnosticsReference.html#%s", s[1], s[2]), nil
}
