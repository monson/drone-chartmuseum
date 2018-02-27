package main

import (
	"io/ioutil"
	"log"

	"github.com/honestbee/drone-chartmuseum/pkg/cmclient"
	"github.com/honestbee/drone-chartmuseum/pkg/util"
)

type (

	// Config struct map with drone plugin parameters
	Config struct {
		RepoURL          string `json:"repo_url,omitempty"`
		ChartPath        string `json:"chart_path,omitempty"`
		ChartDir         string `json:"chart_dir,omitempty"`
		SaveDir          string `json:"save_dir,omitempty"`
		PreviousCommitID string `json:"previous_commit_id,omitempty"`
		CurrentCommitID  string `json:"current_commit_id,omitempty"`
	}

	// Plugin struct
	Plugin struct {
		Config Config
	}
)

// SaveChartToPackage : save helm chart folder to compressed package
func (p *Plugin) SaveChartToPackage(chartPath string) (string, error) {
	var message string
	var err error
	if _, err := os.Stat(p.Config.SaveDir); os.IsNotExist(err) {
		os.Mkdir(p.Config.SaveDir, os.ModePerm)
	}

	if ok, _ := chartutil.IsChartDir(chartPath); ok == true {
		c, _ := chartutil.LoadDir(chartPath)
		message, err = chartutil.Save(c, p.Config.SaveDir)
		if err != nil {
			log.Printf("%v : %v", chartPath, err)
		}
		fmt.Printf("packaging %v ...\n", message)
	}

	return message, err
}

func (p *Plugin) defaultExec(files []string) {
	var resultList []string
	for _, file := range files {
		chart, err := p.SaveChartToPackage(file)
		if err == nil {
			resultList = append(resultList, chart)
		}
	}
	cmclient.UploadToServer(resultList, p.Config.RepoURL)
}

func (p *Plugin) exec() error {
	var files []string
	if p.Config.ChartPath != "" {
		files = []string{p.Config.ChartPath}
	} else if p.Config.PreviousCommitID != "" && p.Config.CurrentCommitID != "" {
		diffFiles := util.GetDiffFiles(p.Config.ChartPath, p.Config.PreviousCommitID, p.Config.CurrentCommitID)
		files = util.GetParentFolders(util.FilterExtFiles(diffFiles))
	} else {
		dirs, err := ioutil.ReadDir(p.Config.ChartDir)
		if err != nil {
			log.Fatal(err)
		}
		files = util.ExtractDirs(dirs)
	}

	p.defaultExec(files)
	return nil
}
