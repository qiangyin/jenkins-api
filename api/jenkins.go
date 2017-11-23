package api

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Auth struct {
	Username string
	ApiToken string
}
type Crumb struct {
	CrumbRequestField string `json:"crumbRequestField"`
	Crumb             string `json:"crumb"`
}

type Jenkins struct {
	auth    *Auth
	baseUrl string
}

func NewJenkins(auth *Auth, baseUrl string) *Jenkins {
	return &Jenkins{
		auth:    auth,
		baseUrl: baseUrl,
	}
}

func (jenkins *Jenkins) buildUrl(path string, params url.Values) (requestUrl string) {
	requestUrl = jenkins.baseUrl + path + "/api/json"
	if params != nil {
		queryString := params.Encode()
		if queryString != "" {
			requestUrl = requestUrl + "?" + queryString
			fmt.Println(requestUrl)
		}
	}

	return
}

func (jenkins *Jenkins) sendRequest(req *http.Request, crumbFlag bool) (*http.Response, error) {
	if crumbFlag {
		crumbUrl := jenkins.buildUrl("/crumbIssuer", nil)
		reqCrumb, _ := http.NewRequest("GET", crumbUrl, nil)
		if jenkins.auth != nil {
			reqCrumb.SetBasicAuth(jenkins.auth.Username, jenkins.auth.ApiToken)
		}
		resp, _ := http.DefaultClient.Do(reqCrumb)
		var crumb Crumb
		jenkins.parseResponse(resp, &crumb)
		req.Header.Add(crumb.CrumbRequestField, crumb.Crumb)
	}
	if jenkins.auth != nil {
		req.SetBasicAuth(jenkins.auth.Username, jenkins.auth.ApiToken)
	}
	return http.DefaultClient.Do(req)
}

func (jenkins *Jenkins) parseXmlResponse(resp *http.Response, body interface{}) (err error) {
	defer resp.Body.Close()

	if body == nil {
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return xml.Unmarshal(data, body)
}

func (jenkins *Jenkins) parseResponse(resp *http.Response, body interface{}) (err error) {
	defer resp.Body.Close()

	if body == nil {
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return json.Unmarshal(data, body)
}

func (jenkins *Jenkins) get(path string, params url.Values, crumbFlag bool, body interface{}) (err error) {
	requestUrl := jenkins.buildUrl(path, params)
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return
	}
	resp, err := jenkins.sendRequest(req, crumbFlag)
	if err != nil {
		return
	}
	return jenkins.parseResponse(resp, body)
}

func (jenkins *Jenkins) getXml(path string, params url.Values, crumbFlag bool, body interface{}) (err error) {
	requestUrl := jenkins.buildUrl(path, params)
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return
	}

	resp, err := jenkins.sendRequest(req, crumbFlag)
	if err != nil {
		return
	}
	return jenkins.parseXmlResponse(resp, body)
}

func (jenkins *Jenkins) post(path string, params url.Values, crumbFlag bool, body interface{}) (err error) {
	requestUrl := jenkins.buildUrl(path, params)
	req, err := http.NewRequest("POST", requestUrl, nil)
	if err != nil {
		return
	}
	resp, err := jenkins.sendRequest(req, crumbFlag)
	if err != nil {
		return
	}
	return jenkins.parseResponse(resp, body)
}
func (jenkins *Jenkins) postXml(path string, params url.Values, xmlBody io.Reader, crumbFlag bool, body interface{}) (err error) {
	requestUrl := jenkins.baseUrl + path
	if params != nil {
		queryString := params.Encode()
		if queryString != "" {
			requestUrl = requestUrl + "?" + queryString
		}
	}

	req, err := http.NewRequest("POST", requestUrl, xmlBody)
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/xml")
	resp, err := jenkins.sendRequest(req, crumbFlag)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("error: HTTP POST returned status code returned: %d", resp.StatusCode))
	}

	return jenkins.parseXmlResponse(resp, body)
}

// GetJobs returns all jobs you can read.
func (jenkins *Jenkins) GetJobs(crumbFlag bool) ([]Job, error) {
	var payload = struct {
		Jobs []Job `json:"jobs"`
	}{}
	err := jenkins.get("", nil, crumbFlag, &payload)
	return payload.Jobs, err
}

// GetJob returns a job which has specified name.
func (jenkins *Jenkins) GetJob(name string, crumbFlag bool) (job Job, err error) {
	err = jenkins.get(fmt.Sprintf("/job/%s", name), nil, crumbFlag, &job)
	return
}

//GetJobConfig returns a maven job, has the one used to create Maven job
func (jenkins *Jenkins) GetJobConfig(name string, crumbFlag bool) (job MavenJobItem, err error) {
	err = jenkins.getXml(fmt.Sprintf("/job/%s/config.xml", name), nil, crumbFlag, &job)
	return
}

// GetBuild returns a number-th build result of specified job.
func (jenkins *Jenkins) GetBuild(job Job, number int, crumbFlag bool) (build Build, err error) {
	err = jenkins.get(fmt.Sprintf("/job/%s/%d", job.Name, number), nil, crumbFlag, &build)
	return
}

// GetLastBuild returns the last build of specified job.
func (jenkins *Jenkins) GetLastBuild(job Job, crumbFlag bool) (build Build, err error) {
	err = jenkins.get(fmt.Sprintf("/job/%s/lastBuild", job.Name), nil, crumbFlag, &build)
	return
}

// Create a new job
func (jenkins *Jenkins) CreateJob(mavenJobItem MavenJobItem, jobName string, crumbFlag bool) error {
	mavenJobItemXml, _ := xml.Marshal(mavenJobItem)
	fmt.Println(string(mavenJobItemXml))
	reader := bytes.NewReader(mavenJobItemXml)
	params := url.Values{"name": []string{jobName}}

	return jenkins.postXml("/createItem", params, reader, crumbFlag, nil)
}

// Create a new job
func (jenkins *Jenkins) CreateJobByCfgXml(cfgXml string, jobName string, crumbFlag bool) error {
	reader := strings.NewReader(cfgXml)
	params := url.Values{"name": []string{jobName}}

	return jenkins.postXml("/createItem", params, reader, crumbFlag, nil)
}

// Add job to view
func (jenkins *Jenkins) AddJobToView(viewName string, job Job, crumbFlag bool) error {
	params := url.Values{"name": []string{job.Name}}
	return jenkins.post(fmt.Sprintf("/view/%s/addJobToView", viewName), params, crumbFlag, nil)
}

// Create a new view
func (jenkins *Jenkins) CreateView(listView ListView, crumbFlag bool) error {
	xmlListView, _ := xml.Marshal(listView)
	reader := bytes.NewReader(xmlListView)
	params := url.Values{"name": []string{listView.Name}}

	return jenkins.postXml("/createView", params, reader, crumbFlag, nil)
}

// Create a new build for this job.
// Params can be nil.
func (jenkins *Jenkins) Build(job Job, params url.Values, crumbFlag bool) error {
	if hasParams(job) {
		return jenkins.post(fmt.Sprintf("/job/%s/buildWithParameters", job.Name), params, crumbFlag, nil)
	} else {
		return jenkins.post(fmt.Sprintf("/job/%s/build", job.Name), params, crumbFlag, nil)
	}
}

// Get the console output from a build.
func (jenkins *Jenkins) GetBuildConsoleOutput(build Build, crumbFlag bool) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/consoleText", build.Url)
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}

	res, err := jenkins.sendRequest(req, crumbFlag)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

// GetQueue returns the current build queue from Jenkins
func (jenkins *Jenkins) GetQueue(crumbFlag bool) (queue Queue, err error) {
	err = jenkins.get(fmt.Sprintf("/queue"), nil, crumbFlag, &queue)
	return
}

// GetArtifact return the content of a build artifact
func (jenkins *Jenkins) GetArtifact(build Build, artifact Artifact, crumbFlag bool) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/artifact/%s", build.Url, artifact.RelativePath)
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}

	res, err := jenkins.sendRequest(req, crumbFlag)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

// SetBuildDescription sets the description of a build
func (jenkins *Jenkins) SetBuildDescription(build Build, description string, crumbFlag bool) error {
	requestUrl := fmt.Sprintf("%ssubmitDescription?description=%s", build.Url, url.QueryEscape(description))
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return err
	}

	res, err := jenkins.sendRequest(req, crumbFlag)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("Unexpected response: expected '200' but received '%d'", res.StatusCode)
	}

	return nil
}

// GetComputerObject returns the main ComputerObject
func (jenkins *Jenkins) GetComputerObject(crumbFlag bool) (co ComputerObject, err error) {
	err = jenkins.get(fmt.Sprintf("/computer"), nil, crumbFlag, &co)
	return
}

// GetComputers returns the list of all Computer objects
func (jenkins *Jenkins) GetComputers(crumbFlag bool) ([]Computer, error) {
	var payload = struct {
		Computers []Computer `json:"computer"`
	}{}
	err := jenkins.get("/computer", nil, crumbFlag, &payload)
	return payload.Computers, err
}

// GetComputer returns a Computer object with a specified name.
func (jenkins *Jenkins) GetComputer(name string, crumbFlag bool) (computer Computer, err error) {
	err = jenkins.get(fmt.Sprintf("/computer/%s", name), nil, crumbFlag, &computer)
	return
}

// hasParams returns a boolean value indicating if the job is parameterized
func hasParams(job Job) bool {
	for _, action := range job.Actions {
		if len(action.ParameterDefinitions) > 0 {
			return true
		}
	}
	return false
}
