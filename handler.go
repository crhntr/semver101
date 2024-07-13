package semver101

import (
	"bytes"
	"context"
	_ "embed"
	"html/template"
	"net/http"
	"strings"

	"github.com/Masterminds/semver/v3"
)

func Handler(prefix string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(res http.ResponseWriter, req *http.Request) {
		inData, ok := newParameters(res, req)
		if !ok {
			return
		}
		outData, err := newResultData(req.Context(), inData, prefix)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		render(res, req, outData, "semver202")
	})
	mux.HandleFunc("GET /results", func(res http.ResponseWriter, req *http.Request) {
		inData, ok := newParameters(res, req)
		if !ok {
			return
		}
		outData, err := newResultData(req.Context(), inData, prefix)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		render(res, req, outData, "result")
	})
	return http.StripPrefix(prefix, mux)
}

type parameters struct {
	Constraints *semver.Constraints `json:"constraint"`
	Versions    []string            `json:"versions"`
}

func newParameters(res http.ResponseWriter, req *http.Request) (parameters, bool) {
	const (
		success = true

		fail = false

		constraintsParamName = "constraints"
		versionsQueryParam   = "versions"
	)
	var (
		constraints = ">= 4.0.6, < 5.3.0"
		versions    = "4.0.6-build.3\n4.0.6\n4.2.0\n5.3.0\n5.3.0\n5.40.0"
	)
	urlQuery := req.URL.Query()
	if urlQuery.Has(constraintsParamName) {
		constraints = urlQuery.Get(constraintsParamName)
	}
	if urlQuery.Has(versionsQueryParam) {
		versions = urlQuery.Get(versionsQueryParam)
	}
	c, err := semver.NewConstraint(constraints)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return parameters{
			Constraints: c,
		}, fail
	}
	versionLines := strings.Split(versions, "\n")
	for i, line := range versionLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		versionLines[i] = strings.TrimSpace(line)
	}
	return parameters{
		Constraints: c,
		Versions:    versionLines,
	}, success
}

type resultDataRow struct {
	Version string   `json:"version"`
	Error   string   `json:"error,omitempty"`
	IsMatch bool     `json:"isMatch,omitempty"`
	Reasons []string `json:"reasons,omitempty"`
}

type resultData struct {
	parameters
	Results    []resultDataRow
	PathPrefix string
}

func newResultData(_ context.Context, in parameters, pathPrefix string) (resultData, error) {
	out := resultData{
		parameters: in,
		PathPrefix: pathPrefix,
	}
	for _, vl := range in.Versions {
		if vl == "" {
			continue
		}
		v, err := semver.NewVersion(vl)
		if err != nil {
			out.Results = append(out.Results, resultDataRow{
				Version: vl,
				Error:   err.Error(),
			})
			continue
		}
		isMatch, reasons := in.Constraints.Validate(v)
		reasonStrings := make([]string, 0, len(reasons))
		for _, r := range reasons {
			reasonStrings = append(reasonStrings, r.Error())
		}
		out.Results = append(out.Results, resultDataRow{
			Version: vl,
			IsMatch: isMatch,
			Reasons: reasonStrings,
		})
	}
	return out, nil
}

var (
	//go:embed index.gohtml
	indexHTML string

	templates = template.Must(template.New("semver202").Parse(indexHTML))
)

func render(res http.ResponseWriter, _ *http.Request, out resultData, templateName string) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024*8))
	if err := templates.ExecuteTemplate(buf, templateName, out); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
	_, _ = res.Write(buf.Bytes())
}
