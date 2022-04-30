package global_value

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
)

// SetAppVersion ---------------------------------------------
func SetAppVersion(appVersion string) {
	_appVersion = appVersion
}

func AppVersion() string {
	return _appVersion
}

// SetExtEnCode ---------------------------------------------
func SetExtEnCode(extEnCode string) {
	_extEnCode = extEnCode
}

func ExtEnCode() string {
	return _extEnCode
}

// ConfigRootDirFPath ---------------------------------------------
func ConfigRootDirFPath() string {

	if _configRootDirFPath == "" {
		_configRootDirFPath = my_folder.GetConfigRootDirFPath()
	}
	return _configRootDirFPath
}

func DefDebugFolder() string {
	var err error
	if _defDebugFolder == "" {
		_defDebugFolder, err = my_folder.GetRootDebugFolder()
		if err != nil {
			panic(err)
		}
	}

	return _defDebugFolder
}

func DefTmpFolder() string {
	var err error
	if _defTmpFolder == "" {
		_defTmpFolder, err = my_folder.GetRootTmpFolder()
		if err != nil {
			panic(err)
		}
	}

	return _defTmpFolder
}

func DefRodTmpRootFolder() string {
	var err error
	if _defRodTmpRootFolder == "" {
		_defRodTmpRootFolder, err = my_folder.GetRodTmpRootFolder()
		if err != nil {
			panic(err)
		}
	}

	return _defRodTmpRootFolder
}

func DefSubFixCacheFolder() string {
	var err error
	if _defSubFixCacheFolder == "" {
		_defSubFixCacheFolder, err = my_folder.GetRootSubFixCacheFolder()
		if err != nil {
			panic(err)
		}
	}

	return _defSubFixCacheFolder
}

func AdblockTmpFolder() string {
	var err error
	if _adblockTmpFolder == "" {
		_adblockTmpFolder, err = my_folder.GetPluginFolderByName(my_folder.Plugin_Adblock)
		if err != nil {
			panic(err)
		}
	}

	return _adblockTmpFolder
}

// ---------------------------------------------
// util.go
var (
	_appVersion           = "" // 程序的版本号
	_extEnCode            = "" // 扩展加密部分
	_configRootDirFPath   = ""
	_defDebugFolder       = ""
	_defTmpFolder         = ""
	_defRodTmpRootFolder  = ""
	_defSubFixCacheFolder = ""
	_adblockTmpFolder     = ""
)
