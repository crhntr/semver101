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

func HandleGet(endpointPath string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		inData, ok := newParameters(res, req)
		if !ok {
			return
		}
		outData, err := newResultData(req.Context(), inData, endpointPath)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		render(res, req, outData, "semver202")
	}
}

func HandlePost(endpointPath string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		inData, ok := newParameters(res, req)
		if !ok {
			return
		}
		outData, err := newResultData(req.Context(), inData, endpointPath)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		render(res, req, outData, "result")
	}
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
	var constraints, versions string
	_ = req.ParseForm()
	constraints = req.Form.Get(constraintsParamName)
	if constraints == "" {
		constraints = "*"
	}
	versions = req.Form.Get(versionsQueryParam)
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
	Results []resultDataRow
	Path    string
}

func newResultData(_ context.Context, in parameters, endpointPath string) (resultData, error) {
	out := resultData{
		parameters: in,
		Path:       endpointPath,
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
