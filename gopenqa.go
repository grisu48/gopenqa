package gopenqa

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

/* Instance defines a openQA instance */
type Instance struct {
	URL           string
	apikey        string
	apisecret     string
	verbose       bool
	maxRecursions int // Maximum number of recursions
}

// the settings are given as a bit of a weird dict:
// e.g. "settings":[{"key":"WORKER_CLASS","value":"\"plebs\""}]}]
// We create an internal struct to account for that
type weirdMachine struct {
	ID       int                 `json:"id"`
	Backend  string              `json:"backend"`
	Name     string              `json:"name"`
	Settings []map[string]string `json:"settings"`
}

// same as weirdMachine for Product
type weirdProduct struct {
	Arch     string              `json:"arch"`
	Distri   string              `json:"distri"`
	Flavor   string              `json:"flavor"`
	Group    string              `json:"group"`
	ID       int                 `json:"id"`
	Version  string              `json:"version"`
	Settings []map[string]string `json:"settings"`
}

func convertSettingsFrom(settings map[string]string) []map[string]string {
	ret := make([]map[string]string, 0)
	for k, v := range settings {
		setting := make(map[string]string, 0)
		setting["key"] = k
		setting["value"] = v
		ret = append(ret, setting)
	}
	return ret
}

func convertSettingsTo(settings []map[string]string) map[string]string {
	ret := make(map[string]string, 0)
	for _, s := range settings {
		k, ok := s["key"]
		if !ok {
			continue
		}
		v, ok := s["value"]
		if !ok {
			continue
		}
		ret[k] = v
	}
	return ret
}

func (mach *weirdMachine) CopySettingsFrom(src Machine) {
	mach.Settings = convertSettingsFrom(src.Settings)
}
func (mach *weirdMachine) CopySettingsTo(dst *Machine) {
	dst.Settings = convertSettingsTo(mach.Settings)
}

func (p *weirdProduct) CopySettingsFrom(src Product) {
	p.Settings = convertSettingsFrom(src.Settings)
}
func (p *weirdProduct) CopySettingsTo(dst *Product) {
	dst.Settings = convertSettingsTo(p.Settings)
}

func (w *weirdProduct) toProduct() Product {
	p := Product{}
	p.Arch = w.Arch
	p.Distri = w.Distri
	p.Flavor = w.Flavor
	p.Group = w.Group
	p.ID = w.ID
	p.Version = w.Version
	w.CopySettingsTo(&p)
	return p
}

func createWeirdProduct(p Product) weirdProduct {
	w := weirdProduct{}
	w.Arch = p.Arch
	w.Distri = p.Distri
	w.Flavor = p.Flavor
	w.Group = p.Group
	w.ID = p.ID
	w.Version = p.Version
	w.CopySettingsFrom(p)
	return w
}

/* Get www-form-urlencoded parameters of this Product */
func (p *weirdProduct) encodeParams() string {
	params := url.Values{}
	params.Add("arch", p.Arch)
	params.Add("distri", p.Distri)
	params.Add("flavor", p.Flavor)
	params.Add("id", fmt.Sprint(p.ID))
	params.Add("version", p.Version)
	for _, s := range p.Settings {
		k, ok := s["key"]
		if !ok {
			continue
		}
		v, ok := s["value"]
		if !ok {
			continue
		}
		params.Add("settings["+k+"]", v)
	}
	return params.Encode()
}

func EmptyParams() map[string]string {
	return make(map[string]string, 0)
}

/* Create a openQA instance module */
func CreateInstance(url string) Instance {
	inst := Instance{URL: url, maxRecursions: 10, verbose: false}
	return inst
}

/* Create a openQA instance module for openqa.opensuse.org */
func CreateO3Instance() Instance {
	return CreateInstance("https://openqa.opensuse.org")
}

// Set the maximum allowed number of recursions before failing
func (i *Instance) SetMaxRecursionDepth(depth int) {
	i.maxRecursions = depth
}

// Set the API key and secret
func (i *Instance) SetApiKey(key string, secret string) {
	i.apikey = key
	i.apisecret = secret
}

func (i *Instance) SetVerbose(verbose bool) {
	i.verbose = verbose
}

func assignInstance(jobs []Job, instance *Instance) []Job {
	for i, j := range jobs {
		j.instance = instance
		jobs[i] = j
	}
	return jobs
}

func hmac_sha1(secret string, key string) []byte {
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(secret))
	return h.Sum(nil)
}

func url_path(url string) string {
	// Ignore http://
	url = strings.Replace(url, "http://", "", 1)
	url = strings.Replace(url, "https://", "", 1)
	// Path from first /
	i := strings.Index(url, "/")
	if i > 0 {
		return url[i:]
	}
	return url
}

/* Perform a GET request on the given url, and send the data as JSON if given
 * Add the APIKEY and APISECRET credentials, if given
 */
func (i *Instance) get(url string, data []byte) ([]byte, error) {
	return i.request("GET", url, data)
}

/* Perform a POST request on the given url, and send the data as JSON if given
 * Add the APIKEY and APISECRET credentials, if given
 */
func (i *Instance) post(url string, data []byte) ([]byte, error) {
	return i.request("POST", url, data)
}

/* Perform a DELETE request on the given url, and send the data as JSON if given
 * Add the APIKEY and APISECRET credentials, if given
 */
func (i *Instance) delete(url string, data []byte) ([]byte, error) {
	return i.request("DELETE", url, data)
}

/* Perform a PUT request on the given url, and send the data as JSON if given
 * Add the APIKEY and APISECRET credentials, if given
 */
func (i *Instance) put(url string, data []byte) ([]byte, error) {
	return i.request("PUT", url, data)
}

/* Perform a request on the given url, and send the data as JSON if given
 * Add the APIKEY and APISECRET credentials, if given
 */
func (i *Instance) request(method string, url string, data []byte) ([]byte, error) {
	contentType := ""
	if data == nil {
		data = make([]byte, 0)
	} else if len(data) > 0 {
		/* Don't do json, but pass it as url encoded form data!
		var err error
		if buf, err = json.Marshal(data); err != nil {
			return buf, err
		}
		*/
		// TODO: Marshall data to URL encoded form data
		contentType = "application/x-www-form-urlencoded"
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return make([]byte, 0), err
	}
	req.Header.Add("Content-Type", contentType)
	// Credentials are sent in the headers
	// "X-API-Key" -> api key
	// "X-API-Hash" -> sha1 hashed api secret
	// POST request
	if i.apikey != "" && i.apisecret != "" {
		req.Header.Add("X-API-Key", i.apikey)
		// The Hash gets salted with the timestamp
		// See https://github.com/os-autoinst/openQA-python-client/blob/master/src/openqa_client/client.py#L115
		// hmac_sha1_sum(/api/v1/machines1617024969, XXXXXXXXXXXXXXXXXX){
		timestamp := time.Now().Unix()
		req.Header.Add("X-API-Microtime", fmt.Sprintf("%d", timestamp))
		path := url_path(url)
		key := fmt.Sprintf("%s%d", path, timestamp)
		hash := fmt.Sprintf("%x", hmac_sha1(key, i.apisecret))
		req.Header.Add("X-API-Hash", hash)

	}
	// Perform request on a new http client
	c := http.Client{}
	r, err := c.Do(req)
	if err != nil {
		return make([]byte, 0), err
	}

	// First read body
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body) // TODO: Limit read size
	if err != nil {
		return buf, err
	}

	// Check status code
	if r.StatusCode != 200 {
		if i.verbose {
			fmt.Fprintf(os.Stderr, "%s\n", string(buf))
		}
		return buf, fmt.Errorf("http status code %d", r.StatusCode)
	}
	return buf, nil
}

/* Query the job overview. params is a map for optional parameters, which will be added to the query.
 * Suitable parameters are `arch`, `distri`, `flavor`, `machine` or `arch`, but everything in this dict will be added to the url
 * Overview returns only the job id and name
 */
func (i *Instance) GetOverview(testsuite string, params map[string]string) ([]Job, error) {
	// Example values:
	// arch=x86_64
	// distri=sle
	// flavor=Server-DVD-Updates
	// machine=64bit

	// Build URL with all parameters
	url := fmt.Sprintf("%s/api/v1/jobs/overview", i.URL)
	if testsuite != "" {
		params["test"] = testsuite
	}
	if len(params) > 0 {
		url += "?" + mergeParams(params)
	}

	jobs, err := fetchJobs(url)
	assignInstance(jobs, i)
	return jobs, err
}

/* Get only the latest jobs of a certain testsuite. Testsuite must be given here.
 * Additional parameters can be supplied via the params map (See GetOverview for more info about usage of those parameters)
 */
func (i *Instance) GetLatestJobs(testsuite string, params map[string]string) ([]Job, error) {
	// Expected result structure
	type ResultJob struct {
		Jobs []Job `json:"jobs"`
	}
	var jobs ResultJob
	if testsuite != "" {
		params["test"] = testsuite
	}
	url := fmt.Sprintf("%s/api/v1/jobs", i.URL)
	if testsuite != "" {
		params["test"] = testsuite
	}
	url += "?" + mergeParams(params)
	// Fetch jobs here, as we expect it to be in `jobs`
	r, err := http.Get(url)
	if err != nil {
		return jobs.Jobs, err
	}
	if r.StatusCode != 200 {
		return jobs.Jobs, fmt.Errorf("http status code %d", r.StatusCode)
	}
	err = json.NewDecoder(r.Body).Decode(&jobs)

	// Now, get only the latest job per group_id
	mapped := make(map[int]Job)
	for _, job := range jobs.Jobs {
		job.instance = i
		// TODO: Filter job results, if given

		// Only keep newer jobs (by ID) per group
		if f, ok := mapped[job.GroupID]; ok {
			if job.ID > f.ID {
				mapped[job.GroupID] = job
			}
		} else {
			mapped[job.GroupID] = job
		}
	}
	// Make slice from map
	ret := make([]Job, 0)
	for _, v := range mapped {
		ret = append(ret, v)
	}
	return ret, nil

}

// GetJob fetches detailled job information
func (i *Instance) GetJob(id int) (Job, error) {
	url := fmt.Sprintf("%s/api/v1/jobs/%d", i.URL, id)
	job, err := fetchJob(url)
	job.Link = fmt.Sprintf("%s/tests/%d", i.URL, id)
	job.instance = i
	return job, err
}

func (i *Instance) DeleteJob(id int) error {
	url := fmt.Sprintf("%s/api/v1/jobs/%d", i.URL, id)
	buf, err := i.delete(url, nil)
	if i.verbose {
		fmt.Fprintf(os.Stderr, "%s\n", buf)
	}
	return err
}

// GetJob fetches detailled job information and follows the job, if it contains a CloneID
func (i *Instance) GetJobFollow(id int) (Job, error) {
	recursions := 0 // keep track of the number of recursions
fetch:
	url := fmt.Sprintf("%s/api/v1/jobs/%d", i.URL, id)
	job, err := fetchJob(url)
	if job.CloneID != 0 && job.CloneID != job.ID {
		recursions++
		if i.maxRecursions != 0 && recursions >= i.maxRecursions {
			return job, fmt.Errorf("maximum recusion depth reached")
		}
		id = job.CloneID
		goto fetch
	}
	job.Link = fmt.Sprintf("%s/tests/%d", i.URL, id)
	job.instance = i
	return job, err
}

func (i *Instance) GetJobGroups() ([]JobGroup, error) {
	url := fmt.Sprintf("%s/api/v1/job_groups", i.URL)
	return fetchJobGroups(url)
}

func (i *Instance) GetJobGroup(id int) (JobGroup, error) {
	url := fmt.Sprintf("%s/api/v1/job_groups/%d", i.URL, id)
	groups, err := fetchJobGroups(url)
	if err != nil {
		return JobGroup{}, err
	}
	if len(groups) == 0 {
		return JobGroup{}, fmt.Errorf("not found")
	}
	return groups[0], nil
}

func (i *Instance) PostJobGroup(jobgroup JobGroup) (JobGroup, error) {
	rurl := fmt.Sprintf("%s/api/v1/job_groups", i.URL)
	//if jobgroup.ID > 0 {
	//	rurl = fmt.Sprintf("%s/api/v1/job_groups/%d", i.URL, jobgroup.ID)
	//}
	buf, err := i.post(rurl, []byte(jobgroup.encodeWWW()))
	if err != nil {
		return jobgroup, err
	}
	err = json.Unmarshal(buf, &jobgroup)
	return jobgroup, err
}

func (i *Instance) GetParentJobGroups() ([]JobGroup, error) {
	url := fmt.Sprintf("%s/api/v1/parent_groups", i.URL)
	return fetchJobGroups(url)
}

func (i *Instance) GetParentJobGroup(id int) (JobGroup, error) {
	url := fmt.Sprintf("%s/api/v1/parent_groups/%d", i.URL, id)
	groups, err := fetchJobGroups(url)
	if err != nil {
		return JobGroup{}, err
	}
	if len(groups) == 0 {
		return JobGroup{}, fmt.Errorf("not found")
	}
	return groups[0], nil
}

func (i *Instance) PostParentJobGroup(jobgroup JobGroup) (JobGroup, error) {
	rurl := fmt.Sprintf("%s/api/v1/parent_groups", i.URL)
	//if jobgroup.ID > 0 {
	//	rurl = fmt.Sprintf("%s/api/v1/parent_groups/%d", i.URL, jobgroup.ID)
	//}
	buf, err := i.post(rurl, []byte(jobgroup.encodeWWW()))
	if err != nil {
		return jobgroup, err
	}
	err = json.Unmarshal(buf, &jobgroup)
	return jobgroup, err
}

func (i *Instance) GetWorkers() ([]Worker, error) {
	url := fmt.Sprintf("%s/api/v1/workers", i.URL)
	return fetchWorkers(url)
}

func fetchJobs(url string) ([]Job, error) {
	jobs := make([]Job, 0)
	r, err := http.Get(url)
	if err != nil {
		return jobs, err
	}
	if r.StatusCode != 200 {
		return jobs, fmt.Errorf("http status code %d", r.StatusCode)
	}
	err = json.NewDecoder(r.Body).Decode(&jobs)
	return jobs, err
}

func fetchJobGroups(url string) ([]JobGroup, error) {
	jobs := make([]JobGroup, 0)
	r, err := http.Get(url)
	if err != nil {
		return jobs, err
	}
	if r.StatusCode != 200 {
		return jobs, fmt.Errorf("http status code %d", r.StatusCode)
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return jobs, err
	}
	// Sometimes SizeLimit is returned as string but it should be an int. Fix this.
	err = json.Unmarshal(data, &jobs)
	return jobs, err
}

func fetchWorkers(url string) ([]Worker, error) {
	r, err := http.Get(url)
	if err != nil {
		return make([]Worker, 0), err
	}
	if r.StatusCode != 200 {
		return make([]Worker, 0), fmt.Errorf("http status code %d", r.StatusCode)
	}
	// workers come in a "workers:[...]" dict
	workers := make(map[string][]Worker, 0)
	err = json.NewDecoder(r.Body).Decode(&workers)
	if workers, ok := workers["workers"]; ok {
		return workers, err
	}
	return make([]Worker, 0), nil
}

func fetchJobTemplates(url string) ([]JobTemplate, error) {
	r, err := http.Get(url)
	if err != nil {
		return make([]JobTemplate, 0), err
	}
	if r.StatusCode != 200 {
		return make([]JobTemplate, 0), fmt.Errorf("http status code %d", r.StatusCode)
	}
	// the templates come as a "JobTemplates:[...]" dict
	templates := make(map[string][]JobTemplate, 0)
	err = json.NewDecoder(r.Body).Decode(&templates)
	if templates, ok := templates["JobTemplates"]; ok {
		return templates, err
	}
	return make([]JobTemplate, 0), nil
}

func fetchMachines(url string) ([]Machine, error) {
	r, err := http.Get(url)
	if err != nil {
		return make([]Machine, 0), err
	}
	if r.StatusCode != 200 {
		return make([]Machine, 0), fmt.Errorf("http status code %d", r.StatusCode)
	}

	// machines come as a "Machines:[...]" dict
	machines := make(map[string][]weirdMachine, 0)
	err = json.NewDecoder(r.Body).Decode(&machines)
	if machines, ok := machines["Machines"]; ok {
		// Parse those weird machines to actual machine instances
		ret := make([]Machine, 0)
		for _, mach := range machines {
			current := Machine{Name: mach.Name, Backend: mach.Backend, ID: mach.ID}
			mach.CopySettingsTo(&current)
			ret = append(ret, current)
		}
		return ret, err
	}
	return make([]Machine, 0), nil
}

func fetchJob(url string) (Job, error) {
	// Expected result structure
	type ResultJob struct {
		Job Job `json:"job"`
	}
	var job ResultJob
	r, err := http.Get(url)
	if err != nil {
		return job.Job, err
	}
	if r.StatusCode != 200 {
		return job.Job, fmt.Errorf("http status code %d", r.StatusCode)
	}
	err = json.NewDecoder(r.Body).Decode(&job)
	return job.Job, err
}

/* merge given parameter string to URL parameters */
func mergeParams(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}
	vals := make([]string, 0)
	for k, v := range params {
		vals = append(vals, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(vals, "&")
}

/*
 * Fetch the given child jobs. Use with j.Children.Chained, j.Children.DirectlyChained and j.Children.Parallel
 * if follow is set to true, the method will return the cloned job instead of the original one, if present
 */
func (j *Job) FetchChildren(children []int, follow bool) ([]Job, error) {
	jobs := make([]Job, 0)
	for _, id := range children {
		job, err := j.instance.GetJobFollow(id)
		if err != nil {
			return jobs, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

/* Fetch all child jobs
 * follow determines if we should follow the given children, i.e. get their cloned jobs instead of the original ones if present
 */
func (j *Job) FetchAllChildren(follow bool) ([]Job, error) {
	children := make([]int, 0)
	children = append(children, j.Children.Chained...)
	children = append(children, j.Children.DirectlyChained...)
	children = append(children, j.Children.Parallel...)
	return j.FetchChildren(children, follow)
}

func (i *Instance) GetJobTemplates() ([]JobTemplate, error) {
	url := fmt.Sprintf("%s/api/v1/job_templates", i.URL)
	return fetchJobTemplates(url)
}

func (instance *Instance) GetJobGroupJobs(id int) ([]int, error) {
	ids := make([]int, 0)
	url := fmt.Sprintf("%s/api/v1/job_groups/%d/jobs", instance.URL, id)
	buf, err := instance.get(url, nil)
	if err != nil {
		return ids, err
	}
	if instance.verbose {
		fmt.Fprintf(os.Stderr, "%s\n", buf)
	}
	var obj map[string][]int // Result: {"ids":[5095,5096,5097,5101,5102]}
	if err = json.Unmarshal(buf, &obj); err != nil {
		return ids, err
	}
	if ids, ok := obj["ids"]; ok {
		return ids, nil
	} else {
		return ids, fmt.Errorf("invalid response")
	}
}

func (i *Instance) DeleteJobGroupJobs(id int) error {
	if jobs, err := i.GetJobGroupJobs(id); err != nil {
		return err
	} else {
		for _, id := range jobs {
			if err := i.DeleteJob(id); err != nil {
				return err
			}
		}
	}
	return nil
}

func (i *Instance) DeleteJobGroup(id int) error {
	url := fmt.Sprintf("%s/api/v1/job_groups/%d", i.URL, id)
	buf, err := i.delete(url, nil)
	if i.verbose {
		fmt.Fprintf(os.Stderr, "%s\n", string(buf))
	}
	return err
}
func (i *Instance) DeleteJobTemplate(id int) error {
	url := fmt.Sprintf("%s/api/v1/job_templates/%d", i.URL, id)
	buf, err := i.delete(url, nil)
	if i.verbose {
		fmt.Fprintf(os.Stderr, "%s\n", string(buf))
	}
	return err
}

func (i *Instance) GetJobTemplate(id int) (JobTemplate, error) {
	url := fmt.Sprintf("%s/api/v1/job_templates/%d", i.URL, id)
	templates, err := fetchJobTemplates(url)
	if err != nil {
		return JobTemplate{}, err
	}
	if len(templates) == 0 {
		return JobTemplate{}, fmt.Errorf("not found")
	} else {
		return templates[0], nil
	}
}

func (i *Instance) GetJobTemplateYAML(id int) (string, error) {
	url := fmt.Sprintf("%s/api/v1/job_templates_scheduling/%d", i.URL, id)
	buf, err := i.get(url, nil)
	return string(buf), err
}
func (i *Instance) PostJobTemplateYAML(id int, yaml string) error {
	url := fmt.Sprintf("%s/api/v1/job_templates_scheduling/%d", i.URL, id)
	_, err := i.post(url, []byte(yaml))
	return err
}

func (i *Instance) GetMachines() ([]Machine, error) {
	url := fmt.Sprintf("%s/api/v1/machines", i.URL)
	return fetchMachines(url)
}

func (i *Instance) GetMachine(id int) (Machine, error) {
	url := fmt.Sprintf("%s/api/v1/machines/%d", i.URL, id)
	if machines, err := fetchMachines(url); err != nil {
		return Machine{}, err
	} else {
		if len(machines) > 0 {
			return machines[0], nil
		} else {
			return Machine{}, nil
		}
	}
}

func (i *Instance) PostMachine(machine Machine) (Machine, error) {
	if i.apikey == "" || i.apisecret == "" {
		return Machine{}, fmt.Errorf("API key or secret not set")
	}

	var rurl string
	if machine.ID == 0 {
		rurl = fmt.Sprintf("%s/api/v1/machines", i.URL)
	} else {
		rurl = fmt.Sprintf("%s/api/v1/machines/%d", i.URL, machine.ID)
	}

	// Add parameters
	params := url.Values{}
	params.Add("backend", machine.Backend)
	params.Add("name", machine.Name)
	for k, v := range machine.Settings {
		params.Add("settings["+k+"]", v)
	}
	rurl += "?" + params.Encode()

	// Setting are encoded in a bit weird way
	// Note: This is not supported by openQA at the moment, but we keep it here for when it does.
	wmach := weirdMachine{Name: machine.Name, ID: machine.ID, Backend: machine.Backend}
	wmach.CopySettingsFrom(machine)

	// Encode the machine settings as JSON
	buf, err := json.Marshal(wmach)
	if err != nil {
		return Machine{}, err
	}
	if buf, err := i.post(rurl, buf); err != nil {
		return Machine{}, err
	} else {
		err = json.Unmarshal(buf, &machine)
		return machine, err
	}
}

func (i *Instance) DeleteMachine(id int) error {
	if i.apikey == "" || i.apisecret == "" {
		return fmt.Errorf("API key or secret not set")
	}

	rurl := fmt.Sprintf("%s/api/v1/machines/%d", i.URL, id)
	buf, err := i.delete(rurl, nil)
	if i.verbose {
		fmt.Fprintf(os.Stderr, "%s\n", string(buf))
	}
	if err != nil {
		return err
	} else {
		return nil
	}
}

func (i *Instance) GetProducts() ([]Product, error) {
	products := make([]Product, 0)
	rurl := fmt.Sprintf("%s/api/v1/products", i.URL)
	buf, err := i.get(rurl, nil)
	if err != nil {
		return products, err
	}
	var obj map[string][]weirdProduct
	if err := json.Unmarshal(buf, &obj); err != nil {
		return products, err
	}
	if fetched, ok := obj["Products"]; ok {
		// Convert from weirdProduct to product
		for _, product := range fetched {
			products = append(products, product.toProduct())
		}
		return products, nil
	}
	if i.verbose {
		fmt.Fprintf(os.Stderr, "%s\n", string(buf))
	}
	return products, fmt.Errorf("invalid response")
}

func (i *Instance) GetProduct(id int) (Product, error) {
	rurl := fmt.Sprintf("%s/api/v1/products/%d", i.URL, id)
	buf, err := i.get(rurl, nil)
	if err != nil {
		return Product{}, err
	}
	var obj map[string][]weirdProduct
	if err := json.Unmarshal(buf, &obj); err != nil {
		return Product{}, err
	}
	if products, ok := obj["Products"]; ok {
		if len(products) == 0 {
			return Product{}, fmt.Errorf("not found")
		}
		return products[0].toProduct(), nil
	} else {
		if i.verbose {
			fmt.Fprintf(os.Stderr, "%s\n", string(buf))
		}
		return Product{}, fmt.Errorf("invalid response")
	}
}

func (i *Instance) PostProduct(product Product) (Product, error) {
	rurl := ""
	if product.ID == 0 {
		rurl = fmt.Sprintf("%s/api/v1/products", i.URL)
	} else {
		rurl = fmt.Sprintf("%s/api/v1/products/%d", i.URL, product.ID)
	}

	// Product to values
	wproduct := createWeirdProduct(product)
	data := []byte(wproduct.encodeParams())
	if i.verbose {
		fmt.Fprintf(os.Stderr, "%s\n", data)
	}
	buf, err := i.post(rurl, data)
	if i.verbose {
		fmt.Fprintf(os.Stderr, "%s\n", string(buf))
	}
	if err != nil {
		return Product{}, err
	}
	err = json.Unmarshal(buf, &product)
	return product, err
}

/* Fetch comments for a given job */
func (i *Instance) GetComments(job int) ([]Comment, error) {
	ret := make([]Comment, 0)
	rurl := fmt.Sprintf("%s/api/v1/jobs/%d/comments", i.URL, job)
	buf, err := i.get(rurl, nil)
	if i.verbose {
		fmt.Fprintf(os.Stderr, "%s\n", string(buf))
	}
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(buf, &ret)
	return ret, err
}
