package backend

import (
	"fmt"
	"os"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"

	"github.com/ChineseSubFinder/ChineseSubFinder/cmd/GetCAPTCHA/backend/config"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/something_static"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/sirupsen/logrus"
	sshOrg "golang.org/x/crypto/ssh"
)

func GitProcess(log *logrus.Logger, config config.Config, enString string) error {

	log.Infoln("Now Time", time.Now().Format("2006-01-02 15:04:05"))
	var delFileNames []string
	nowTT := time.Now()
	nowTime := nowTT.Format("2006-01-02")
	nowTimeFileNamePrix := fmt.Sprintf("%d%d%d", nowTT.Year(), nowTT.Month(), nowTT.Day())

	// 实例化登录密钥
	publicKeys, err := ssh.NewPublicKeysFromFile("git", config.SSHKeyFullPath, config.SSHKeyPwd)
	if err != nil {
		return err
	}
	publicKeys.HostKeyCallback = sshOrg.InsecureIgnoreHostKey()

	var r *git.Repository
	var w *git.Worktree
	if pkg.IsDir(config.CloneProjectDesSaveDir) == true {
		// 需要 pull
		log.Infoln("Pull Start...")
		r, err = git.PlainOpen(config.CloneProjectDesSaveDir)
		if err != nil {
			return err
		}
		w, err = r.Worktree()
		if err != nil {
			return err
		}
		err = w.Pull(&git.PullOptions{Auth: publicKeys, RemoteName: "origin"})
		if err != nil {
			if err != git.NoErrAlreadyUpToDate {
				return err
			}
		}
		log.Infoln("Pull End")

	} else {
		// 需要 clone
		log.Infoln("PlainClone Start...")

		r, err = git.PlainClone(config.CloneProjectDesSaveDir, false, &git.CloneOptions{
			Auth:     publicKeys,
			URL:      config.GitProjectUrl,
			Progress: os.Stdout,
		})
		if err != nil {
			return err
		}
		log.Infoln("PlainClone End")

	}
	// 存储外部传入的字符串到文件
	bok, err := something_static.WriteFile(config.CloneProjectDesSaveDir, enString, nowTime, nowTimeFileNamePrix)
	if err != nil {
		return err
	}
	if bok == false {
		// 说明无需继续，因为文件没有变化
		log.Infoln("Code not change, Skip This Time")
		return nil
	}

	log.Infoln("Write File Done")
	w, err = r.Worktree()
	if err != nil {
		return err
	}

	// 遍历当前文件夹，仅仅保留当天的文件 nowTimeFileNamePrix + common.StaticFileName00
	delFileNames, err = delExpireFile(log, config.CloneProjectDesSaveDir, nowTimeFileNamePrix+common.StaticFileName00)
	if err != nil {
		return err
	}
	for _, delFileName := range delFileNames {
		_, err = w.Remove(delFileName)
		if err != nil {
			return err
		}
	}
	// 添加文件到 git 文件树
	_, err = w.Add(nowTimeFileNamePrix + common.StaticFileName00)
	if err != nil {
		return err
	}
	status, err := w.Status()
	log.Infoln("Status", status)
	commit, err := w.Commit("update", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "haha",
			Email: "haha@haha.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}
	_, err = r.CommitObject(commit)
	if err != nil {
		return err
	}
	log.Infoln("Commit Done")
	err = r.Push(&git.PushOptions{
		Auth: publicKeys,
	})
	if err != nil {
		return err
	}

	log.Infoln("Push Done.")

	return nil
}

func delExpireFile(log *logrus.Logger, dir string, goldName string) ([]string, error) {

	delFileNames := make([]string, 0)
	pathSep := string(os.PathSeparator)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, curFile := range files {
		fullPath := dir + pathSep + curFile.Name()
		if curFile.IsDir() {
			continue
		} else {

			if "ReadMe.md" == curFile.Name() {
				continue
			}
			// 这里就是文件了
			if curFile.Name() != goldName {

				log.Infoln("Del Expire File:", fullPath)
				err = os.Remove(fullPath)
				if err != nil {
					return nil, err
				}
				delFileNames = append(delFileNames, curFile.Name())
			}
		}
	}

	return delFileNames, nil
}
