package pkg

import (
	"strconv"
	"strings"
)

// SetAppVersion ---------------------------------------------
func SetAppVersion(appVersion string) {
	_appVersion = appVersion
}

func AppVersion() string {
	return _appVersion
}

func AppVersionInt() (int, int, int) {

	if _major == 0 && _minor == 0 && _patch == 0 {

		nowAppVersion := strings.ToLower(_appVersion)
		nowAppVersion = strings.ReplaceAll(nowAppVersion, "v", "")
		if strings.Contains(nowAppVersion, "-lite") == true {
			nowAppVersion = strings.ReplaceAll(nowAppVersion, "-lite", "")
		}
		if strings.Contains(nowAppVersion, "-beta") == true {
			nowAppVersion = strings.Split(nowAppVersion, "-beta")[0]
		}
		versions := strings.Split(nowAppVersion, ".")
		if len(versions) == 3 {
			_major, _ = strconv.Atoi(versions[0])
			_minor, _ = strconv.Atoi(versions[1])
			_patch, _ = strconv.Atoi(versions[2])
			return _major, _minor, _patch
		} else {
			return 0, 0, 0
		}
	}

	return _major, _minor, _patch
}

// SetExtEnCode ---------------------------------------------
func SetExtEnCode(extEnCode string) {
	_extEnCode = extEnCode
}

func ExtEnCode() string {
	return _extEnCode
}

// SetBaseKey ---------------------------------------------
func SetBaseKey(baseKey string) {
	_baseKey = baseKey
}

func BaseKey() string {
	return _baseKey
}

// SetAESKey16 ---------------------------------------------
func SetAESKey16(aESKey16 string) {
	_aESKey16 = aESKey16
}

func AESKey16() string {
	return _aESKey16
}

// SetAESIv16 ---------------------------------------------
func SetAESIv16(aESIv16 string) {
	_aESIv16 = aESIv16
}

func AESIv16() string {
	return _aESIv16
}

// ConfigRootDirFPath ---------------------------------------------
func ConfigRootDirFPath() string {

	if _configRootDirFPath == "" {
		_configRootDirFPath = GetConfigRootDirFPath()
	}
	return _configRootDirFPath
}

func DefDebugFolder() string {
	var err error
	if _defDebugFolder == "" {
		_defDebugFolder, err = GetRootDebugFolder()
		if err != nil {
			panic(err)
		}
	}

	return _defDebugFolder
}

func DefTmpFolder() string {
	var err error
	if _defTmpFolder == "" {
		_defTmpFolder, err = GetRootTmpFolder()
		if err != nil {
			panic(err)
		}
	}

	return _defTmpFolder
}

func DefRodTmpRootFolder() string {
	var err error
	if _defRodTmpRootFolder == "" {
		_defRodTmpRootFolder, err = GetRodTmpRootFolder()
		if err != nil {
			panic(err)
		}
	}

	return _defRodTmpRootFolder
}

func DefSubFixCacheFolder() string {
	var err error
	if _defSubFixCacheFolder == "" {
		_defSubFixCacheFolder, err = GetRootSubFixCacheFolder()
		if err != nil {
			panic(err)
		}
	}

	return _defSubFixCacheFolder
}

func AdblockTmpFolder() string {
	var err error
	if _adblockTmpFolder == "" {
		_adblockTmpFolder, err = GetPluginFolderByName(Plugin_Adblock)
		if err != nil {
			panic(err)
		}
	}

	return _adblockTmpFolder
}

// LiteMode ---------------------------------------------
func LiteMode() bool {
	return _liteMode
}

func SetLiteMode(liteMode bool) {
	_liteMode = liteMode
}

// LinuxConfigPathInSelfPath ---------------------------------------------
// 针对制作群晖的 SPK 应用，无法写入默认的 /config 目录而给出的新的编译条件，直接指向这个目录到当前程序的目录
func LinuxConfigPathInSelfPath() string {
	return setLinuxConfigPathInSelfPath
}

func SetLinuxConfigPathInSelfPath(setPath string) {

	setLinuxConfigPathInSelfPath = setPath
}

var setLinuxConfigPathInSelfPath = ""

// ---------------------------------------------
// util.go
var (
	_appVersion           = "" // 程序的版本号
	_extEnCode            = "" // 扩展加密部分
	_baseKey              = "" // 基础的密钥，密钥会基于这个基础的密钥生成
	_aESKey16             = "" // AES密钥
	_aESIv16              = "" // 初始化向量
	_configRootDirFPath   = ""
	_defDebugFolder       = ""
	_defTmpFolder         = ""
	_defRodTmpRootFolder  = ""
	_defSubFixCacheFolder = ""
	_adblockTmpFolder     = ""
	_liteMode             = false
	_major                = 0
	_minor                = 0
	_patch                = 0
)
