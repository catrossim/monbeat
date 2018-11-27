package command

import (
	"bytes"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/elastic/beats/libbeat/logp"

	"github.com/catrossim/monbeat/utils"

	"github.com/elastic/beats/libbeat/common"
)

type CmdWatcher struct {
	cmd      string
	internal time.Duration
	cache    string
	out      chan *common.MapStr
	err      chan error
	lock     sync.Mutex
	logger   *logp.Logger
}

func NewCmdWatcher(cmd string, internal time.Duration, out chan *common.MapStr, logger *logp.Logger, err chan error) (*CmdWatcher, error) {
	return &CmdWatcher{
		cmd:      cmd,
		internal: internal,
		out:      out,
		err:      err,
		logger:   logger,
	}, nil
}

func (cw *CmdWatcher) Watch(done chan struct{}) error {
	ticker := time.NewTicker(cw.internal)
	cw.logger.Debug("start command watcher")
	// watch periodlly
	for {
		select {
		case <-done:
			return nil
		case <-ticker.C:
		}

		result, err := execCmd(cw.cmd)
		if err != nil {
			cw.reportErr(err)
			continue
		}
		md5Token, err := utils.GenFileToken(result)
		if err != nil {
			cw.reportErr(err)
			continue
		}
		if cw.cache == "" {
			cw.logger.Debugf("First execution of %s", cw.cmd)
			cw.cache = md5Token
			output := &common.MapStr{
				"cmd":     cw.cmd,
				"md5":     md5Token,
				"content": string(result),
			}
			cw.out <- output
		} else {
			prevToken := cw.cache
			currToken := md5Token
			if prevToken != currToken {
				// send result
				output := &common.MapStr{
					"cmd":     cw.cmd,
					"md5":     md5Token,
					"content": string(result),
				}
				cw.out <- output
			}
			cw.cache = md5Token
		}
	}
}

func execCmd(command string) ([]byte, error) {
	tokens := strings.Split(command, " ")
	cmd := exec.Command(tokens[0], tokens[1:]...)
	var out bytes.Buffer
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	uid, err := strconv.Atoi(usr.Uid)
	if err != nil {
		panic(err)
	}

	gid, err := strconv.Atoi(usr.Gid)
	if err != nil {
		panic(err)
	}
	cmd.Stdout = &out
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid:         uint32(uid),
			Gid:         uint32(gid),
			NoSetGroups: true,
		},
	}
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func (cw *CmdWatcher) reportErr(err error) {
	cw.logger.Error(err)
	cw.err <- err
}
