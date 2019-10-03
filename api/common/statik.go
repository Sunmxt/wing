package common

import (
	"github.com/rakyll/statik/fs"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"

	_ "git.stuhome.com/Sunmxt/wing/statik"
)

var StaticsFS http.FileSystem

func init() {
	s, err := fs.New()
	if err != nil {
		log.Panic("[statics] failed to load static http filesystem: ", err.Error())
	}
	StaticsFS = s
}

func TryOpenStaticFile(rctx *RequestContext, path string, defaultPath string) (file http.File, stat os.FileInfo, err error) {
	serveDefault := false
	setDefault := func() {
		rctx.OpCtx.Log.Infof("[statics] serve index \"%v\" for path \"%v\".", defaultPath, path)
		serveDefault = true
		path = defaultPath
		file = nil
		stat = nil
	}

	for {
		file, err = StaticsFS.Open(path)
		if err != nil {
			if err == os.ErrNotExist && !serveDefault {
				setDefault()
				continue
			}
			return nil, nil, err
		}
		if serveDefault {
			return file, stat, err
		}
		if file == nil {
			setDefault()
			continue
		}
		stat, err = file.Stat()
		if stat.IsDir() || err != nil {
			setDefault()
			continue
		}
		break
	}
	return file, stat, nil
}
