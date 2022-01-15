package pre_download_process

import "errors"

type PreDownloadProcess struct {
	stageName string
	gError    error
}

func NewPreDownloadProcess() *PreDownloadProcess {
	return &PreDownloadProcess{}
}

func (p *PreDownloadProcess) Init() *PreDownloadProcess {

	if p.gError != nil {
		return p
	}
	p.stageName = "Init"
	// do something

	return p
}

func (p *PreDownloadProcess) Start() *PreDownloadProcess {

	if p.gError != nil {
		return p
	}
	p.stageName = "Start"
	// do something

	return p
}

func (p *PreDownloadProcess) Do() error {

	return errors.New(p.stageName + " " + p.gError.Error())
}
