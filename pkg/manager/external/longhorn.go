package external

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/harvester/support-bundle-utils/pkg/utils"
)

const (
	BundleStateReadyForDownload = "ReadyForDownload"
)

type LonghornSBManager struct {
	context    context.Context
	backendURL string
	httpClient *http.Client
}

type LonghornSupportBundleInitateInput struct {
	IssueURL    string `json:"issueURL"`
	Description string `json:"description"`
}

type Resource struct {
	ID      string            `json:"id,omitempty"`
	Type    string            `json:"type,omitempty"`
	Links   map[string]string `json:"links"`
	Actions map[string]string `json:"actions"`
}

type LonghornSupportBundle struct {
	Resource
	NodeID             string `json:"nodeID"`
	State              string `json:"state"`
	Name               string `json:"name"`
	ErrorMessage       string `json:"errorMessage"`
	ProgressPercentage int    `json:"progressPercentage"`
}

func NewLonghornSupportBundleManager(ctx context.Context, backendURL string) *LonghornSBManager {
	return &LonghornSBManager{
		context:    ctx,
		backendURL: backendURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (m *LonghornSBManager) GetLonghornSupportBundle(issueURL string, description string, toDir string) error {
	bundleInput := LonghornSupportBundleInitateInput{
		IssueURL:    issueURL,
		Description: description,
	}

	sb, err := m.createBundle(bundleInput)
	if err != nil {
		return err
	}

	err = m.waitBundle(sb)
	if err != nil {
		return err
	}

	err = m.downloadBundle(sb, toDir)
	if err != nil {
		return err
	}
	return nil
}

func (m *LonghornSBManager) createBundle(input LonghornSupportBundleInitateInput) (*LonghornSupportBundle, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1/supportbundles", m.backendURL)
	resp, err := m.httpReq(http.MethodPost, url, payload)
	if err != nil {
		return nil, err
	}

	var lsb LonghornSupportBundle
	err = json.Unmarshal(resp, &lsb)
	if err != nil {
		return nil, err
	}
	return &lsb, nil
}

func (m *LonghornSBManager) waitBundle(sb *LonghornSupportBundle) error {
	interval := 5 * time.Second
	timeout := 2 * time.Minute

	readyCondition := func() (done bool, err error) {
		newSb, err := m.getBundle(sb.NodeID, sb.Name)
		if err != nil {
			return false, err
		}
		if newSb.State == BundleStateReadyForDownload {
			return true, nil
		}
		return false, nil
	}

	err := wait.PollImmediate(interval, timeout, readyCondition)
	if err != nil {
		return errors.New("timeout for waiting a bundle")
	}
	return nil
}

func (m *LonghornSBManager) getBundle(node string, bundleName string) (*LonghornSupportBundle, error) {
	url := fmt.Sprintf("%s/v1/supportbundles/%s/%s", m.backendURL, node, bundleName)
	resp, err := m.httpReq(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var lsb LonghornSupportBundle
	err = json.Unmarshal(resp, &lsb)
	if err != nil {
		return nil, err
	}
	return &lsb, nil
}

func (m *LonghornSBManager) downloadBundle(sb *LonghornSupportBundle, toDir string) error {
	url := fmt.Sprintf("%s/v1/supportbundles/%s/%s/download", m.backendURL, sb.NodeID, sb.Name)
	req, err := http.NewRequestWithContext(m.context, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	filename, err := utils.HttpGetDispositionFilename(resp.Header.Get("Content-Disposition"))
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(toDir, filename))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	logrus.Debugf("file %s is downloaded to %s", filename, toDir)
	return err
}

func (m *LonghornSBManager) httpReq(method string, url string, data []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(m.context, method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}
