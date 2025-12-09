package templates

import (
	_ "embed"
)

//go:embed assets/style.css
var cssData string

//go:embed assets/main.js
var jsData string

//go:embed assets/success.css
var successCSSData string

//go:embed assets/success.js
var successJSData string

//go:embed assets/docs.css
var docsCSSData string

//go:embed assets/docs.js
var docsJSData string

func css() string {
	return cssData
}

func javascript() string {
	return jsData
}

func successCSS() string {
	return successCSSData
}

func successJS() string {
	return successJSData
}

func docsCSS() string {
	return docsCSSData
}

func docsJS() string {
	return docsJSData
}
