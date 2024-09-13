package service

import (
	"errors"
	"github.com/nathanmbicho/savannahtech/filetracker/app/internal/service/filetrack"
)

var ErrNoEmptyFilePath = errors.New("models: no file path provided")
var ErrNoFile = errors.New("models: no such existing file record found")

type Service struct {
	FileTracker filetrack.FileTracker
}

func NewService() Service {
	return Service{
		FileTracker: filetrack.FileTracker{},
	}
}
