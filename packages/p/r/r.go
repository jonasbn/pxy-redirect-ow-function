package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Response struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

const tpl = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <title>pxy.fi</title>
    <link
      rel="stylesheet"
      href="https://cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.1.3/css/bootstrap.min.css"
    />
    <link
      rel="stylesheet"
      href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.11.2/css/all.min.css"
    />
    <link
      rel="stylesheet"
      href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome-animation/0.2.1/font-awesome-animation.min.css"
    />
    <style>
      body {
        background: #000e29;
      }

      .alert > .start-icon {
        margin-right: 0;
        min-width: 20px;
        text-align: center;
      }

      .alert > .start-icon {
        margin-right: 5px;
      }

      .greencross {
        font-size: 18px;
        color: #25ff0b;
        text-shadow: none;
      }

      .alert-simple.alert-success {
        border: 1px solid rgba(36, 241, 6, 0.46);
        background-color: rgba(7, 149, 66, 0.12156862745098039);
        box-shadow: 0px 0px 2px #259c08;
        color: #0ad406;
        text-shadow: 2px 1px #00040a;
        transition: 0.5s;
        cursor: pointer;
      }
      .alert-success:hover {
        background-color: rgba(7, 149, 66, 0.35);
        transition: 0.5s;
      }
      .alert-simple.alert-info {
        border: 1px solid rgba(6, 44, 241, 0.46);
        background-color: rgba(7, 73, 149, 0.12156862745098039);
        box-shadow: 0px 0px 2px #0396ff;
        color: #0396ff;
        text-shadow: 2px 1px #00040a;
        transition: 0.5s;
        cursor: pointer;
      }

      .alert-info:hover {
        background-color: rgba(7, 73, 149, 0.35);
        transition: 0.5s;
      }

      .blue-cross {
        font-size: 18px;
        color: #0bd2ff;
        text-shadow: none;
      }

      .alert-simple.alert-warning {
        border: 1px solid rgba(241, 142, 6, 0.81);
        background-color: rgba(220, 128, 1, 0.16);
        box-shadow: 0px 0px 2px #ffb103;
        color: #ffb103;
        text-shadow: 2px 1px #00040a;
        transition: 0.5s;
        cursor: pointer;
      }

      .alert-warning:hover {
        background-color: rgba(220, 128, 1, 0.33);
        transition: 0.5s;
      }

      .warning {
        font-size: 18px;
        color: #ffb40b;
        text-shadow: none;
      }

      .alert-danger>p>a {
        color: #ff0303;
      }

      .alert-simple.alert-danger {
        border: 1px solid rgba(241, 6, 6, 0.81);
        background-color: rgba(220, 17, 1, 0.16);
        box-shadow: 0px 0px 2px #ff0303;
        color: #ff0303;
        text-shadow: 2px 1px #00040a;
        transition: 0.5s;
        cursor: pointer;
      }

      .alert-danger:hover {
        background-color: rgba(220, 17, 1, 0.33);
        transition: 0.5s;
      }

      .danger {
        font-size: 18px;
        color: #ff0303;
        text-shadow: none;
      }

      .alert-simple.alert-primary {
        border: 1px solid rgba(6, 241, 226, 0.81);
        background-color: rgba(1, 204, 220, 0.16);
        box-shadow: 0px 0px 2px #03fff5;
        color: #03d0ff;
        text-shadow: 2px 1px #00040a;
        transition: 0.5s;
        cursor: pointer;
      }

      .alert-primary:hover {
        background-color: rgba(1, 204, 220, 0.33);
        transition: 0.5s;
      }

      .alertprimary {
        font-size: 18px;
        color: #03d0ff;
        text-shadow: none;
      }

      .square_box {
        position: absolute;
        -webkit-transform: rotate(45deg);
        -ms-transform: rotate(45deg);
        transform: rotate(45deg);
        border-top-left-radius: 45px;
        opacity: 0.302;
      }

      .square_box.box_three {
        background-image: -moz-linear-gradient(
          -90deg,
          #290a59 0%,
          #3d57f4 100%
        );
        background-image: -webkit-linear-gradient(
          -90deg,
          #290a59 0%,
          #3d57f4 100%
        );
        background-image: -ms-linear-gradient(-90deg, #290a59 0%, #3d57f4 100%);
        opacity: 0.059;
        left: -80px;
        top: -60px;
        width: 500px;
        height: 500px;
        border-radius: 45px;
      }

      .square_box.box_four {
        background-image: -moz-linear-gradient(
          -90deg,
          #290a59 0%,
          #3d57f4 100%
        );
        background-image: -webkit-linear-gradient(
          -90deg,
          #290a59 0%,
          #3d57f4 100%
        );
        background-image: -ms-linear-gradient(-90deg, #290a59 0%, #3d57f4 100%);
        opacity: 0.059;
        left: 150px;
        top: -25px;
        width: 550px;
        height: 550px;
        border-radius: 45px;
      }

      .alert:before {
        content: "";
        position: absolute;
        width: 0;
        height: calc(100% - 44px);
        border-left: 1px solid;
        border-right: 2px solid;
        border-bottom-right-radius: 3px;
        border-top-right-radius: 3px;
        left: 0;
        top: 50%;
        transform: translate(0, -50%);
        height: 20px;
      }

      .fa-times {
        -webkit-animation: blink-1 2s infinite both;
        animation: blink-1 2s infinite both;
      }

      .my-times {
        -webkit-animation: blink-1 2s infinite both;
	      animation: blink-1 2s infinite both;
      }

      /**
    * ----------------------------------------
    * animation blink-1
    * ----------------------------------------
    */
      @-webkit-keyframes blink-1 {
        0%,
        50%,
        100% {
          opacity: 1;
        }
        25%,
        75% {
          opacity: 0;
        }
      }
      @keyframes blink-1 {
        0%,
        50%,
        100% {
          opacity: 1;
        }
        25%,
        75% {
          opacity: 0;
        }
      }
    </style>
  </head>
  <body>
    <!-- partial:index.partial.html -->
    <section>
      <div class="square_box box_three"></div>
      <div class="square_box box_four"></div>
      <div class="container mt-5">
        <div class="row">
		{{if eq .PageType "error"}}
		<div class="col-sm-12">
		<div
		class="alert fade alert-simple alert-danger alert-dismissible text-left font__family-montserrat font__size-16 font__weight-light brk-library-rendered rendered show"
		role="alert"
		data-brk-library="component__alert"
		>
		<i class="start-icon far fa-times-circle faa-pulse animated"></i>
		<strong class="font__weight-semibold">Error</strong> <p>{{ .Message }}</p>
		</div>
		</div>
		{{else if eq .PageType "warning"}}
		<div class="col-sm-12">
		<div
		class="alert fade alert-simple alert-warning alert-dismissible text-left font__family-montserrat font__size-16 font__weight-light brk-library-rendered rendered show"
		role="alert"
		data-brk-library="component__alert"
		>
		<i
			class="start-icon fa fa-exclamation-triangle faa-flash animated"
		></i>
		<strong class="font__weight-semibold">Warning</strong> <p>{{ .Message }}</p>
		</div>
		</div>
		{{else}}
		<div class="col-sm-12">
		<div
		  class="alert fade alert-simple alert-info alert-dismissible text-left font__family-montserrat font__size-16 font__weight-light brk-library-rendered rendered show"
		  role="alert"
		  data-brk-library="component__alert"
		>
		  <i class="start-icon fa fa-info-circle faa-shake animated"></i>
		  <strong class="font__weight-semibold">Welcome</strong> {{ .Message }}
		</div>
		</div>
		{{end}}
        </div>
      </div>
    </section>
    <!-- partial -->
  </body>
</html>`

const info = `
<p>This is <b>pxy-redirect-ow-function</b> served at:<br><br><b>pxy.fi - /ˈpɒksify/</b>
<p>The purpose of pxy-redirect-ow-function is too offer short links for a page which broke when a set of links got too long, this is a <i>really <a href="https://dev.to/jonasbn/challenges-solutions-and-more-challenges-and-more-solutions-4j3f">long story</a></i>.</p>
<p>For more documentation and information please see links below</p>
<br>
<p>The layout for this page is based on a Codepen by Swarup Kumar Kuila</p>

<ul>
<li><a href="https://github.com/jonasbn/pxy-redirect-ow-function">GitHub</a></li>
<li><a href="https://codepen.io/uiswarup/pen/RwNraeW">Codepen</a></li>
<li><a href="https://jonasbn.github.io/">Author</a></li>
</ul>
`

type TmplData struct {
	Message  template.HTML
	PageType string
}

func Main(args map[string]interface{}) *Response {

	log.SetLevel(log.InfoLevel)

	if os.Getenv("LOGLEVEL") != "" {
		if os.Getenv("LOGLEVEL") == "debug" {
			log.SetLevel(log.DebugLevel)
		}
	}

	userAgent := ""
	ip := ""
	referer := ""
	path := ""

	if args["__ow_path"].(string) != "" {
		path = args["__ow_path"].(string)
	}

	requestHeaders := args["__ow_headers"]
	val, _ := requestHeaders.(map[string]interface{})

	if val["user-agent"].(string) != "" {
		userAgent = val["user-agent"].(string)
	}

	if val["do-connecting-ip"].(string) != "" {
		ip = val["do-connecting-ip"].(string)
	}

	if val["referer"].(string) != "" {
		referer = val["referer"].(string)
	}

	url, err := parseRedirectURL(path, ip, userAgent, referer)

	if err != nil {

		b, renderErr := renderPage(err.Error(), "error")

		if renderErr != nil {
			return &Response{
				StatusCode: http.StatusInternalServerError,
				Body:       renderErr.Error(),
			}
		}

		return &Response{
			StatusCode: http.StatusInternalServerError,
			Body:       b.String(),
		}
	}

	if url.String() == "/" || url.String() == "/index.html" {
		log.Infof("Non-redirectable URL >%s< served", url.String())

		b, renderErr := renderPage(info, "info")

		if renderErr != nil {
			return &Response{
				StatusCode: http.StatusInternalServerError,
				Body:       renderErr.Error(),
			}
		}

		return &Response{
			StatusCode: http.StatusOK,
			Body:       b.String(),
		}
	}

	targetURL, err := redirect(url)

	if err != nil {

		b, renderErr := renderPage(err.Error(), "error")

		if renderErr != nil {
			return &Response{
				StatusCode: http.StatusInternalServerError,
				Body:       renderErr.Error(),
			}
		}

		return &Response{
			StatusCode: http.StatusBadRequest,
			Body:       b.String(),
		}
	}

	headers := make(map[string]string)
	headers["location"] = targetURL

	log.Infof("Redirecting to: >%s<", targetURL)

	return &Response{
		Headers:    headers,
		StatusCode: http.StatusTemporaryRedirect,
	}
}

func renderPage(message string, pagetype string) (bytes.Buffer, error) {

	var b bytes.Buffer

	tmpl, err := template.New("").Parse(tpl)
	if err != nil {
		log.Errorf("Unable to parse HTML template: %s", err.Error())
		return b, err
	}

	data := TmplData{Message: template.HTML(message), PageType: pagetype}

	err = tmpl.Execute(&b, &data)

	if err != nil {
		log.Errorf("Unable to render HTML page: %s", err.Error())
		return b, err
	}

	return b, nil
}

func parseRedirectURL(path, ip, userAgent, referer string) (*url.URL, error) {

	log.WithFields(log.Fields{
		"ip":         ip,
		"user-agent": userAgent,
		"referer":    referer,
	}).Info("Received URL: >" + path + "< via >" + referer)

	redirectURL, parseErr := url.Parse(path)
	if parseErr != nil {
		log.Errorf("Unable to parse received URL: >%s<", path)
		return nil, fmt.Errorf("Unable to parse received URL: >%s<", path)
	}

	log.Debugf("Parsed URL: >%s<", redirectURL)

	return redirectURL, nil
}

func redirect(url *url.URL) (string, error) {

	redirectURL, err := assembleRedirectURL(url)

	if err != nil {
		return "", err
	}

	return redirectURL, nil
}

func assembleRedirectURL(url *url.URL) (string, error) {

	s := strings.SplitN(url.Path, "/", 3)

	log.Debugf("Parsed following parts: >%#v<", s)

	// 0 is empty because we split on "/" and the URL begins with "/"
	// 1 == version
	// 2 == fragment

	url.Host = "pxy.fi"
	url.Scheme = "https"

	if len(s) < 3 {
		log.Errorf("insufficient parts in provided url: >%s<", url.String())

		// Example:
		// https://pxy.fi/p/r/5
		err := fmt.Errorf("<p>You only made it this far, because the specified URL has insufficient parts to redirect to the documentation</p><p>%s://%s/p/r/<span class=\"my-times\">%s<span></p><p>In order to get the redirect to work, please specify both a version and a fragment</p><p>Example: https://pxy.fi/p/r/13/wall</p><p>See more information at: <a href=\"https://github.com/jonasbn/pxy-redirect-ow-function\">GitHub</a></p>", url.Scheme, url.Host, s[1])

		return "", err
	}

	_, err := strconv.Atoi(s[1])
	if err != nil {
		log.Errorf("first part of url: >%s< is not a number: %q", url.String(), s)

		// Example:
		// https://pxy.fi/p/r/X
		err := fmt.Errorf("<p>You only made it this far, because the specified URL requires a version number as the first part to redirect to the documentation</p><p>%s://%s/p/r/<span class=\"my-times\">%s</span>/%s</p><p>In order to get the redirect to work, please specify both a version and a fragment</p><p>Example: https://pxy.fi/p/r/13/wall</p><p>See more information at: <a href=\"https://github.com/jonasbn/pxy-redirect-ow-function\">GitHub</a></p>", url.Scheme, url.Host, s[1], s[2])
		return "", err
	}

	if s[2] == "" {
		log.Errorf("second part of url: >%s< is not a string: %q", url.String(), s)

		// Example:
		// Example: https://pxy.fi/p/r/0
		err := fmt.Errorf("<p>You only made it this far, because the specified URL requires a part to indicatede the fragment as the second part to redirect to the documentation</p>%s://%s/p/r/%s/<span class=\"my-times\">%s</span></p><p>In order to get the redirect to work, please specify both a version and a fragment</p><p>Example: https://pxy.fi/p/r/13/wall</p><p>See more information at: <a href=\"https://github.com/jonasbn/pxy-redirect-ow-function\">GitHub</a></p>", url.Scheme, url.Host, s[1], "x")
		return "", err
	}

	return fmt.Sprintf("https://releases.llvm.org/%s.0.0/tools/clang/docs/DiagnosticsReference.html#%s", s[1], s[2]), nil
}
