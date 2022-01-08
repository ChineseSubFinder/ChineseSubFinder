package backend

import (
	"github.com/allanpk716/ChineseSubFinder/cmd/GetCAPTCHA/backend/config"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/something_static"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	sshOrg "golang.org/x/crypto/ssh"
	"os"
	"time"
)

func GitProcess(config config.Config, enString string) error {

	log_helper.GetLogger().Infoln("Now Time", time.Now().Format("2006-01-02 15:04:05"))
	nowTime := time.Now().Format("2006-01-02")

	// 实例化登录密钥
	publicKeys, err := ssh.NewPublicKeysFromFile("git", config.SSHKeyFullPath, config.SSHKeyPwd)
	if err != nil {
		return err
	}
	publicKeys.HostKeyCallback = sshOrg.InsecureIgnoreHostKey()

	var r *git.Repository
	var w *git.Worktree
	if my_util.IsDir(config.CloneProjectDesSaveDir) == true {
		// 需要 pull
		log_helper.GetLogger().Infoln("Pull Start...")
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
		log_helper.GetLogger().Infoln("Pull End")
	} else {
		// 需要 clone
		log_helper.GetLogger().Infoln("PlainClone Start...")

		r, err = git.PlainClone(config.CloneProjectDesSaveDir, false, &git.CloneOptions{
			Auth:     publicKeys,
			URL:      config.GitProjectUrl,
			Progress: os.Stdout,
		})
		if err != nil {
			return err
		}
		log_helper.GetLogger().Infoln("PlainClone End")

	}
	// 存储外部传入的字符串到文件
	bok, err := something_static.WriteFile(config.CloneProjectDesSaveDir, enString, nowTime)
	if err != nil {
		return err
	}
	if bok == false {
		// 说明无需继续，因为文件没有变化
		log_helper.GetLogger().Infoln("Code not change, Skip This Time")
		return nil
	}

	log_helper.GetLogger().Infoln("Write File Done")
	w, err = r.Worktree()
	if err != nil {
		return err
	}
	_, err = w.Add(common.StaticFileName00)
	if err != nil {
		return err
	}
	status, err := w.Status()
	log_helper.GetLogger().Infoln("Status", status)
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
	log_helper.GetLogger().Infoln("Commit Done")
	err = r.Push(&git.PushOptions{
		Auth: publicKeys,
	})
	if err != nil {
		return err
	}

	log_helper.GetLogger().Infoln("Push Done.")

	return nil
}
