package backend

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/cmd/GetCAPTCHA/backend/config"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
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

	fmt.Println("Now Time", time.Now().Format("2006-01-02 15:04:05"))
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
		fmt.Println("Pull Start...")
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
		fmt.Println("Pull End")
	} else {
		// 需要 clone
		fmt.Println("PlainClone Start...")

		r, err = git.PlainClone(config.CloneProjectDesSaveDir, false, &git.CloneOptions{
			Auth:     publicKeys,
			URL:      config.GitProjectUrl,
			Progress: os.Stdout,
		})
		if err != nil {
			return err
		}
		fmt.Println("PlainClone End")

	}
	// 存储外部传入的字符串到文件
	bok, err := something_static.WriteFile(config.CloneProjectDesSaveDir, enString, nowTime)
	if err != nil {
		return err
	}
	if bok == false {
		// 说明无需继续，因为文件没有变化
		fmt.Println("Code not change, Skip This Time")
		return nil
	}

	fmt.Println("Write File Done")
	w, err = r.Worktree()
	if err != nil {
		return err
	}
	_, err = w.Add(common.StaticFileName00)
	if err != nil {
		return err
	}
	status, err := w.Status()
	fmt.Println("Status", status)
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
	fmt.Println("Commit Done")
	err = r.Push(&git.PushOptions{
		Auth: publicKeys,
	})
	if err != nil {
		return err
	}

	fmt.Println("Push Done.")

	return nil
}
