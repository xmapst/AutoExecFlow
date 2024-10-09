package git

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	ghttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"

	"github.com/xmapst/AutoExecFlow/internal/runner/common"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

type Git struct {
	ctx       context.Context
	cncl      context.CancelFunc
	gopt      *git.CloneOptions
	conf      *gitConfig
	storage   storage.IStep
	workspace string
}

type gitConfig struct {
	Url          string `json:"url"`
	User         string `json:"user"`
	Token        string `json:"token"`
	SshKey       string `json:"ssh_key"`
	SshKeyPaas   string `json:"ssh_key_pass"`
	Branch       string `json:"branch"`
	Sha          string `json:"sha"`
	SubDir       string `json:"sub_dir"`
	SingleBranch *bool  `json:"single_branch,omitempty"`
}

func New(storage storage.IStep, workspace string) (*Git, error) {
	g := &Git{
		gopt: &git.CloneOptions{
			SingleBranch: true,
		},
		conf: &gitConfig{
			SubDir: "",
			User:   "git",
		},
		storage:   storage,
		workspace: workspace,
	}
	g.conf.SshKey = filepath.Join(g.homeDir(), ".ssh", "id_rsa")
	if err := g.parseConfig(); err != nil {
		return nil, err
	}
	return g, nil
}

func (g *Git) parseConfig() error {
	content, err := g.storage.Content()
	if err != nil {
		return err
	}
	if err = json.Unmarshal([]byte(content), g.conf); err != nil {
		return err
	}
	g.gopt.URL = g.conf.Url
	g.workspace = filepath.Join(g.workspace, g.conf.SubDir)
	if g.conf.SingleBranch != nil {
		g.gopt.SingleBranch = *g.conf.SingleBranch
	}
	if g.conf.Sha == "" {
		g.conf.Sha = g.conf.Branch
	}
	if g.conf.Token != "" {
		if g.conf.User != "" {
			g.gopt.Auth = &ghttp.BasicAuth{
				Username: g.conf.User,
				Password: g.conf.Token,
			}
		} else {
			g.gopt.Auth = &ghttp.TokenAuth{
				Token: g.conf.Token,
			}
		}
	} else {
		// 尝试打开文件
		keyBts, err := os.ReadFile(g.conf.SshKey)
		if err != nil {
			// return nil, err
			// 没有文件就直接使用key
			keyBts = []byte(g.conf.SshKey)
		}
		pk, err := gssh.NewPublicKeys(g.conf.User, keyBts, g.conf.SshKeyPaas)
		if err != nil {
			return err
		}
		pk.HostKeyCallback = ssh.InsecureIgnoreHostKey()
		g.gopt.Auth = pk
	}
	if g.gopt.URL == "" {
		return errors.New("git urls is empty")
	}
	if g.conf.Branch != "" {
		g.gopt.ReferenceName = plumbing.NewBranchReferenceName(g.conf.Branch)
	} else if g.conf.Sha != "" && !plumbing.IsHash(g.conf.Sha) {
		g.gopt.ReferenceName = plumbing.NewBranchReferenceName(g.conf.Sha)
	}
	if plumbing.IsHash(g.conf.Sha) {
		g.gopt.SingleBranch = false
	}
	return nil
}

func (g *Git) homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		if runtime.GOOS == "windows" {
			home = os.Getenv("USERPROFILE")
		} else {
			home = os.Getenv("HOME")
		}
	}
	return home
}

func (g *Git) Run(ctx context.Context) (code int64, err error) {
	logx.Infoln("git clone")
	g.storage.Log().Write("git clone")
	timeout, err := g.storage.Timeout()
	if err != nil {
		return common.CodeSystemErr, err
	}
	g.ctx, g.cncl = context.WithCancel(ctx)
	if timeout > 0 {
		g.ctx, g.cncl = context.WithTimeout(ctx, timeout)
	}
	var cmdout = &bytes.Buffer{}
	g.gopt.Progress = cmdout
	rpy, err := git.PlainCloneContext(g.ctx, g.workspace, false, g.gopt)
	if err != nil {
		return common.CodeSystemErr, fmt.Errorf("cloneRepo err:%v", err)
	}
	if plumbing.IsHash(g.conf.Sha) {
		worktree, err := rpy.Worktree()
		if err != nil {
			return common.CodeSystemErr, err
		}
		err = worktree.Checkout(&git.CheckoutOptions{
			Force: true,
			Hash:  plumbing.NewHash(g.conf.Sha),
		})
		if err != nil {
			return common.CodeSystemErr, fmt.Errorf("CheckOutHash [%s] err:%v", g.conf.Sha, err)
		}
	}
	g.storage.Log().Write(cmdout.String())
	return common.CodeSuccess, err
}

func (g *Git) Clear() error {
	if g.cncl != nil {
		g.cncl()
		g.cncl = nil
	}
	return nil
}
