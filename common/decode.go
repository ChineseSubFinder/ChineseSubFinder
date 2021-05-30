package common

import (
	"github.com/beevik/etree"
	"io/ioutil"
	"os"
	"strings"
)

func get_IMDB_movie_xml(movieFilePath string) (string, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(movieFilePath); err != nil {
		return "", err
	}
	for _, t := range doc.FindElements("//IMDB") {
		return t.Text(), nil
	}

	return "", CanNotFindIMDBID
}

func get_IMDB_nfo(nfoFilePath string) (string, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(nfoFilePath); err != nil {
		return "", err
	}
	for _, t := range doc.FindElements("//uniqueid[@type='Imdb']") {
		return t.Text(), nil
	}

	return "", CanNotFindIMDBID
}

func Get_IMDB_Id(dirPth string) (string ,error) {
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return "", err
	}
	pathSep := string(os.PathSeparator)
	// 优先找 movie.xml 这个是 raddarr 下载的电影会存下来的，可以在 Metadata 设置 Emby
	var movieFilePath = ""
	// 这个是使用 tinyMediaManager 削刮器按 Kodi 来存储的
	var nfoFilePath = ""

	for _, fi := range dir {
		if fi.IsDir() == true {
			continue
		}
		upperName := strings.ToUpper(fi.Name())
		// 找 movie.xml
		if upperName == strings.ToUpper(metadataFileEmby) {
			movieFilePath = dirPth + pathSep + fi.Name()
		}
		// 找 *.nfo
		ok := strings.HasSuffix(fi.Name(), suffixNameNfo)
		if ok {
			nfoFilePath = dirPth + pathSep + fi.Name()
		}
	}
	// 根据找到的开始解析
	if movieFilePath == "" && nfoFilePath == "" {
		return "", NoMetadataFile
	}

	if movieFilePath != "" {
		outId, err := get_IMDB_movie_xml(movieFilePath)
		if err != nil {
			println(err)
		} else {
			return outId, nil
		}
	}

	if nfoFilePath != "" {
		outId, err := get_IMDB_nfo(nfoFilePath)
		if err != nil {
			return "", err
		} else {
			return outId, nil
		}
	}

	return "", CanNotFindIMDBID
}

const (
	metadataFileEmby = "movie.xml"
	suffixNameXml    = ".xml"
	suffixNameNfo    = ".nfo"
)