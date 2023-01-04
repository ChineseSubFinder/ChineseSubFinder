package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/get_access_time"
	"github.com/sirupsen/logrus"
)

// --------------------------------------------------------------
// Debug
// --------------------------------------------------------------

// GetRootDebugFolder 在程序的根目录新建，调试用文件夹
func GetRootDebugFolder() (string, error) {

	nowProcessRoot, err := os.Getwd()
	if err != nil {
		return "", err
	}
	nowProcessRoot = filepath.Join(nowProcessRoot, cacheRootFolderName, DebugFolder)
	err = os.MkdirAll(nowProcessRoot, os.ModePerm)
	if err != nil {
		return "", err
	}

	return nowProcessRoot, err
}

// GetDebugFolderByName 根据传入的 strings (["aa", "bb"]) 组成  DebugFolder/aa/bb 这样的路径
func GetDebugFolderByName(names []string) (string, error) {

	rootPath, err := GetRootDebugFolder()
	if err != nil {
		return "", err
	}

	tmpFolderFullPath := rootPath
	for _, name := range names {
		tmpFolderFullPath = filepath.Join(tmpFolderFullPath, name)
	}
	err = os.MkdirAll(tmpFolderFullPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	return tmpFolderFullPath, nil
}

// CopyFiles2DebugFolder 把文件放入到 Debug 文件夹中，新建 desFolderName 文件夹
func CopyFiles2DebugFolder(names []string, subFiles []string) error {
	debugFolderByName, err := GetDebugFolderByName(names)
	if err != nil {
		return err
	}

	// 复制下载在 tmp 文件夹中的字幕文件到视频文件夹下面
	for _, subFile := range subFiles {
		newFn := filepath.Join(debugFolderByName, filepath.Base(subFile))
		err = CopyFile(subFile, newFn)
		if err != nil {
			return err
		}
	}

	return nil
}

// --------------------------------------------------------------
// Tmp
// --------------------------------------------------------------

// GetRootTmpFolder 在程序的根目录新建，取缓用文件夹，每一个视频的缓存将在其中额外新建子集文件夹
func GetRootTmpFolder() (string, error) {

	nowProcessRoot, err := os.Getwd()
	if err != nil {
		return "", err
	}
	nowProcessRoot = filepath.Join(nowProcessRoot, cacheRootFolderName, TmpFolder)
	err = os.MkdirAll(nowProcessRoot, os.ModePerm)
	if err != nil {
		return "", err
	}
	return nowProcessRoot, err
}

// GetTmpFolderByName 获取缓存的文件夹，没有则新建
func GetTmpFolderByName(folderName string) (string, error) {
	rootPath, err := GetRootTmpFolder()
	if err != nil {
		return "", err
	}
	tmpFolderFullPath := filepath.Join(rootPath, folderName)
	err = os.MkdirAll(tmpFolderFullPath, os.ModePerm)
	if err != nil {
		return "", err
	}
	return tmpFolderFullPath, nil
}

// ClearTmpFolderByName 清理指定的缓存文件夹
func ClearTmpFolderByName(folderName string) error {

	nowTmpFolder, err := GetTmpFolderByName(folderName)
	if err != nil {
		return err
	}

	return ClearFolder(nowTmpFolder)
}

// ClearRootTmpFolder 清理缓存的根目录，将里面的子文件夹一并清理
func ClearRootTmpFolder() error {
	nowTmpFolder, err := GetRootTmpFolder()
	if err != nil {
		return err
	}

	pathSep := string(os.PathSeparator)
	files, err := os.ReadDir(nowTmpFolder)
	if err != nil {
		return err
	}
	for _, curFile := range files {
		fullPath := nowTmpFolder + pathSep + curFile.Name()
		if curFile.IsDir() {
			err = os.RemoveAll(fullPath)
			if err != nil {
				return err
			}
		} else {
			// 这里就是文件了
			err = os.Remove(fullPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// --------------------------------------------------------------
// Adblock Cache
// --------------------------------------------------------------

// GetPluginRootFolder 在程序的根目录新建，取缓用文件夹，每一个视频的缓存将在其中额外新建子集文件夹
func GetPluginRootFolder() (string, error) {

	nowProcessRoot, err := os.Getwd()
	if err != nil {
		return "", err
	}
	nowProcessRoot = filepath.Join(nowProcessRoot, cacheRootFolderName, PluginFolder)
	err = os.MkdirAll(nowProcessRoot, os.ModePerm)
	if err != nil {
		return "", err
	}
	return nowProcessRoot, err
}

// GetPluginFolderByName 获取缓存的文件夹，没有则新建
func GetPluginFolderByName(folderName string) (string, error) {
	rootPath, err := GetPluginRootFolder()
	if err != nil {
		return "", err
	}
	tmpFolderFullPath := filepath.Join(rootPath, folderName)
	err = os.MkdirAll(tmpFolderFullPath, os.ModePerm)
	if err != nil {
		return "", err
	}
	return tmpFolderFullPath, nil
}

// ClearPluginFolderByName 清理指定的缓存文件夹
func ClearPluginFolderByName(folderName string) error {

	nowTmpFolder, err := GetPluginFolderByName(folderName)
	if err != nil {
		return err
	}

	return ClearFolder(nowTmpFolder)
}

// --------------------------------------------------------------
// Rod Cache
// --------------------------------------------------------------

// GetRodTmpRootFolder 在程序的根目录新建，rod 缓存用文件夹
func GetRodTmpRootFolder() (string, error) {

	nowProcessRoot, err := os.Getwd()
	if err != nil {
		return "", err
	}
	nowProcessRoot = filepath.Join(nowProcessRoot, cacheRootFolderName, RodCacheFolder)
	err = os.MkdirAll(nowProcessRoot, os.ModePerm)
	if err != nil {
		return "", err
	}
	return nowProcessRoot, err
}

// ClearRodTmpRootFolder 清理 rod 缓存文件夹
func ClearRodTmpRootFolder() error {

	nowTmpFolder, err := GetRodTmpRootFolder()
	if err != nil {
		return err
	}

	pathSep := string(os.PathSeparator)
	files, err := os.ReadDir(nowTmpFolder)
	if err != nil {
		return err
	}
	for _, curFile := range files {
		fullPath := nowTmpFolder + pathSep + curFile.Name()
		if curFile.IsDir() {
			err = os.RemoveAll(fullPath)
			if err != nil {
				return err
			}
		} else {
			// 这里就是文件了
			err = os.Remove(fullPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// --------------------------------------------------------------
// Sub Fix Cache
// --------------------------------------------------------------

// GetRootSubFixCacheFolder 在程序的根目录新建，字幕时间校正的缓存文件夹
func GetRootSubFixCacheFolder() (string, error) {

	nowProcessRoot, err := os.Getwd()
	if err != nil {
		return "", err
	}
	nowProcessRoot = filepath.Join(nowProcessRoot, cacheRootFolderName, SubFixCacheFolder)
	err = os.MkdirAll(nowProcessRoot, os.ModePerm)
	if err != nil {
		return "", err
	}
	return nowProcessRoot, err
}

// GetRootCacheCenterFolder 下载缓存、队列缓存、下载次数缓存的文件夹
func GetRootCacheCenterFolder() (string, error) {

	nowProcessRoot, err := os.Getwd()
	if err != nil {
		return "", err
	}
	nowProcessRoot = filepath.Join(nowProcessRoot, cacheRootFolderName, CacheCenterFolder)
	err = os.MkdirAll(nowProcessRoot, os.ModePerm)
	if err != nil {
		return "", err
	}
	return nowProcessRoot, err
}

// GetSubFixCacheFolderByName 获取缓存的文件夹，没有则新建
func GetSubFixCacheFolderByName(folderName string) (string, error) {
	rootPath, err := GetRootSubFixCacheFolder()
	if err != nil {
		return "", err
	}
	tmpFolderFullPath := filepath.Join(rootPath, folderName)
	err = os.MkdirAll(tmpFolderFullPath, os.ModePerm)
	if err != nil {
		return "", err
	}
	return tmpFolderFullPath, nil
}

// --------------------------------------------------------------
// Share Sub Cache
// --------------------------------------------------------------

// GetShareSubRootFolder 在程序的根目录新建，字幕共享的缓存根目录，下级还有具体是按发行的时间去划分的子集目录
func GetShareSubRootFolder() (string, error) {

	nowProcessRoot, err := os.Getwd()
	if err != nil {
		return "", err
	}
	nowProcessRoot = filepath.Join(nowProcessRoot, cacheRootFolderName, ShareSubFileCache)
	err = os.MkdirAll(nowProcessRoot, os.ModePerm)
	if err != nil {
		return "", err
	}
	return nowProcessRoot, err
}

// GetShareFolderByYear 缓存的文件夹以发行的年为一个单位存储
func GetShareFolderByYear(year int) (string, error) {
	rootPath, err := GetShareSubRootFolder()
	if err != nil {
		return "", err
	}
	tmpFolderFullPath := filepath.Join(rootPath, fmt.Sprintf("%d", year))
	err = os.MkdirAll(tmpFolderFullPath, os.ModePerm)
	if err != nil {
		return "", err
	}
	return tmpFolderFullPath, nil
}

// ClearShareSubFolderByYear 清理指定的缓存文件夹
func ClearShareSubFolderByYear(year int) error {

	nowTmpFolder, err := GetShareFolderByYear(year)
	if err != nil {
		return err
	}

	return ClearFolder(nowTmpFolder)
}

// ClearShareSubFolderByYearAndName 清理指定的缓存文件夹
func ClearShareSubFolderByYearAndName(year int, name string) error {

	nowTmpFolder, err := GetShareFolderByYear(year)
	if err != nil {
		return err
	}

	return ClearFolder(filepath.Join(nowTmpFolder, name))
}

// --------------------------------------------------------------
// Manual Subtitle Upload
// --------------------------------------------------------------

// GetManualSubUploadCacheFolder 手动上传字幕的缓存文件夹
func GetManualSubUploadCacheFolder() (string, error) {

	nowProcessRoot, err := os.Getwd()
	if err != nil {
		return "", err
	}
	nowProcessRoot = filepath.Join(nowProcessRoot, cacheRootFolderName, ManualSubUploadCacheFolder)
	err = os.MkdirAll(nowProcessRoot, os.ModePerm)
	if err != nil {
		return "", err
	}
	return nowProcessRoot, err
}

// ClearManualSubUploadCacheFolder 清理手动上传字幕的缓存文件夹
func ClearManualSubUploadCacheFolder() error {

	nowTmpFolder, err := GetManualSubUploadCacheFolder()
	if err != nil {
		return err
	}

	pathSep := string(os.PathSeparator)
	files, err := os.ReadDir(nowTmpFolder)
	if err != nil {
		return err
	}
	for _, curFile := range files {
		fullPath := nowTmpFolder + pathSep + curFile.Name()
		if curFile.IsDir() {
			err = os.RemoveAll(fullPath)
			if err != nil {
				return err
			}
		} else {
			// 这里就是文件了
			err = os.Remove(fullPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// --------------------------------------------------------------
// 视频和字幕的预览缓存
// --------------------------------------------------------------

// GetVideoAndSubPreviewCacheFolder 视频和字幕的预览缓存
func GetVideoAndSubPreviewCacheFolder() (string, error) {

	nowProcessRoot, err := os.Getwd()
	if err != nil {
		return "", err
	}
	nowProcessRoot = filepath.Join(nowProcessRoot, cacheRootFolderName, VideoAndSubPreviewCacheFolder)
	err = os.MkdirAll(nowProcessRoot, os.ModePerm)
	if err != nil {
		return "", err
	}
	return nowProcessRoot, err
}

// ClearVideoAndSubPreviewCacheFolder 清理视频和字幕的预览缓存
func ClearVideoAndSubPreviewCacheFolder() error {

	nowTmpFolder, err := GetVideoAndSubPreviewCacheFolder()
	if err != nil {
		return err
	}

	pathSep := string(os.PathSeparator)
	files, err := os.ReadDir(nowTmpFolder)
	if err != nil {
		return err
	}
	for _, curFile := range files {
		fullPath := nowTmpFolder + pathSep + curFile.Name()
		if curFile.IsDir() {
			err = os.RemoveAll(fullPath)
			if err != nil {
				return err
			}
		} else {
			// 这里就是文件了
			err = os.Remove(fullPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// --------------------------------------------------------------
// Common
// --------------------------------------------------------------

// ClearFolder 清空文件夹
func ClearFolder(folderFullPath string) error {
	pathSep := string(os.PathSeparator)
	files, err := os.ReadDir(folderFullPath)
	if err != nil {
		return err
	}
	for _, curFile := range files {
		fullPath := folderFullPath + pathSep + curFile.Name()
		if curFile.IsDir() {
			err = os.RemoveAll(fullPath)
			if err != nil {
				return err
			}
		} else {
			// 这里就是文件了
			err = os.Remove(fullPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetConfigRootDirFPath 获取 Config 的根目录，不同系统不一样
func GetConfigRootDirFPath() string {

	nowConfigFPath := ""
	sysType := runtime.GOOS
	if sysType == "linux" {

		if LinuxConfigPathInSelfPath() != "" {
			// 自定义路径
			nowConfigFPath = LinuxConfigPathInSelfPath()
		} else {
			// 专用目录，与 Docker 设置一致
			nowConfigFPath = configDirRootFPathLinux
		}

	} else if sysType == "windows" {
		if LinuxConfigPathInSelfPath() != "" {
			// 自定义路径
			nowConfigFPath = LinuxConfigPathInSelfPath()
		} else {
			nowConfigFPath = configDirRootFPathWindows
		}
	} else if sysType == "darwin" {

		if LinuxConfigPathInSelfPath() != "" {
			// 自定义路径
			nowConfigFPath = LinuxConfigPathInSelfPath()
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				panic("GetConfigRootDirFPath darwin get UserHomeDir, Error:" + err.Error())
			}
			nowConfigFPath = home + configDirRootFPathDarwin
		}
	} else {
		panic("GetConfigRootDirFPath can't matched OSType: " + sysType + " ,You Should Implement It Yourself")
	}

	return nowConfigFPath
}

// ClearIdleSubFixCacheFolder 清理闲置的字幕修正缓存文件夹
func ClearIdleSubFixCacheFolder(l *logrus.Logger, rootSubFixCacheFolder string, outOfDate time.Duration) error {

	/*
		从 GetRootSubFixCacheFolder 目录下，遍历第一级目录中的文件夹
		然后每个文件夹中，统计里面最后的访问时间（可能有多个文件），如果超过某个时间范围就标记删除这个文件夹
	*/
	pathSep := string(os.PathSeparator)
	files, err := os.ReadDir(rootSubFixCacheFolder)
	if err != nil {
		return err
	}
	wait2ScanFolder := make([]string, 0)
	for _, curFile := range files {

		fullPath := rootSubFixCacheFolder + pathSep + curFile.Name()
		if curFile.IsDir() == true {
			// 需要关注文件夹
			wait2ScanFolder = append(wait2ScanFolder, fullPath)
		}
	}

	wait2DeleteFolder := make([]string, 0)
	getAccessTimeEx := get_access_time.GetAccessTimeEx{}
	cutOff := time.Now().Add(-outOfDate)
	for _, s := range wait2ScanFolder {

		files, err = os.ReadDir(s)
		if err != nil {
			return err
		}

		maxAccessTime := time.Now()
		// 需要统计这个文件夹下的所有文件的 AccessTIme，找出最新（最大的值）的那个时间，再比较
		for i, curFile := range files {

			fullPath := s + pathSep + curFile.Name()
			if curFile.IsDir() == true {
				continue
			}
			// 只需要关注文件
			accessTime, err := getAccessTimeEx.GetAccessTime(fullPath)
			if err != nil {
				return err
			}
			if i == 0 {
				maxAccessTime = accessTime
			}
			if Time2SecondNumber(accessTime) > Time2SecondNumber(maxAccessTime) {
				maxAccessTime = accessTime
			}
		}
		if maxAccessTime.Sub(cutOff) <= 0 {
			// 确认可以删除
			wait2DeleteFolder = append(wait2DeleteFolder, s)
		}
	}
	// 统一清理过期的文件夹
	for _, s := range wait2DeleteFolder {
		l.Infoln("Try 2 clear SubFixCache Folder:", s)
		err := os.RemoveAll(s)
		if err != nil {
			return err
		}
	}

	return nil
}

// 缓存文件的位置信息，都是在程序的根目录下的 cache 中
const (
	cacheRootFolderName           = "cache"                       // 缓存文件夹总名称
	TmpFolder                     = "tmp"                         // 临时缓存的文件夹
	RodCacheFolder                = "rod"                         // rod 的缓存目录
	PluginFolder                  = "Plugin"                      // 插件的目录
	DebugFolder                   = "CSF-DebugThings"             // 调试相关的文件夹
	SubFixCacheFolder             = "CSF-SubFixCache"             // 字幕时间校正的缓存文件夹，一般可以不清理
	ShareSubFileCache             = "CSF-ShareSubCache"           // 字幕共享的缓存目录，不建议删除
	CacheCenterFolder             = "CSF-CacheCenter"             // 下载缓存、队列缓存、下载次数缓存的文件夹
	ManualSubUploadCacheFolder    = "CSF-ManualSubUploadCache"    // 手动上传字幕的缓存文件夹
	VideoAndSubPreviewCacheFolder = "CSF-VideoAndSubPreviewCache" // 视频和字幕的预览缓存
)

const (
	Plugin_Adblock = "adblock"
)

// 配置文件的位置信息，这个会根据系统版本做区分
const (
	configDirRootFPathWindows = "."                         // Windows 就是在当前的程序目录
	configDirRootFPathLinux   = "/config"                   // Linux 是在 /config 下
	configDirRootFPathDarwin  = "/.config/chinesesubfinder" // Darwin 是在 os.UserHomeDir()/.config/chinesesubfinder/ 下
)
