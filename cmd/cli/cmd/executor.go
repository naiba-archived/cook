package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"

	"github.com/p14yground/cook/dao"
	"github.com/p14yground/cook/model"
	"github.com/p14yground/cook/pkg/log"
)

const (
	endLabel = "cook-exec-label:done"
)

// NewExecutor ..
func NewExecutor() *Executor {
	var e Executor
	e.sessions = make(map[string]*sshRW)
	return &e
}

type sshRW struct {
	Stdout *bytes.Buffer
	Stdin  io.Writer
}

// Executor ..
type Executor struct {
	sessions map[string]*sshRW
	servers  []*model.Server
}

// Exec ..
func (e *Executor) Exec(s string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return
	} else if s == "quit" || s == "exit" {
		fmt.Println("Bye!")
		os.Exit(0)
		return
	}
	var err error
	args := strings.Split(s, " ")

	switch args[0] {
	case "connect":
		if len(args) >= 2 {
			switch args[1] {
			case "--all":
				e.connect(nil, true)
				return
			case "--tags":
				if len(args) != 3 {
					break
				}
				e.connect(strings.Split(args[2], ","), false)
				return
			}
		}
		err = errors.New("参数不完整")
	case "exec":
		command := s[4:]
		if len(command) == 0 {
			err = errors.New("空命令")
			break
		}
		e.run(command)
		return
	}

	log.Printf("错误命令：%s %v", s, err)
}

func (e *Executor) run(command string) {
	if len(e.sessions) == 0 {
		log.Printf("没有建立的连接")
		return
	}
	var wg sync.WaitGroup
	wg.Add(len(e.sessions))
	for host, session := range e.sessions {
		go func(host string, session *sshRW) {
			var err error
			allDone := make(chan struct{})
			defer func() {
				<-allDone
				wg.Done()
			}()
			log.Printf("在 %s 中执行 %s", host, command)
			stdoutClose := make(chan struct{})
			go func() {
				out := e.readStdout(session.Stdout, stdoutClose)
				log.Printf("------- [%s] log -------\n%s------- [%s] -------", host, out, host)
				close(allDone)
			}()
			_, err = fmt.Fprintf(session.Stdin, "%s && echo %s \n", command, endLabel)
			if err != nil {
				log.Printf("在 %s 中执行时出现问题：%v", host, err)
				close(stdoutClose)
			}
		}(host, session)
	}
	wg.Wait()
}

func (e *Executor) connect(tags []string, isAll bool) {
	if isAll {
		for _, server := range dao.Servers {
			e.servers = e.merge(e.servers, server)
		}
	} else {
		for i := 0; i < len(tags); i++ {
			if ss, has := dao.Servers[tags[i]]; has {
				e.servers = e.merge(e.servers, ss)
			} else {
				log.Printf("Tag：%s, 不存在", tags[i])
			}
		}
	}

	var labels []string
	for i := 0; i < len(e.servers); i++ {
		labels = append(labels, e.servers[i].Label)
	}

	log.Printf("即将连接 %d 个服务器：%v", len(e.servers), labels)
	var wg sync.WaitGroup
	wg.Add(len(e.servers))
	for i := 0; i < len(e.servers); i++ {
		go func(i int) {
			defer wg.Done()
			addr := fmt.Sprintf("%s:%s", e.servers[i].Host, e.servers[i].Port)
			config := &ssh.ClientConfig{
				Timeout:         time.Second * 5, //ssh 连接time out 时间一秒钟, 如果ssh验证错误 会在一秒内返回
				User:            e.servers[i].User,
				HostKeyCallback: ssh.InsecureIgnoreHostKey(), //这个可以， 但是不够安全
				//HostKeyCallback: hostKeyCallBackFunc(h.Host),
			}
			if e.servers[i].Password != "" {
				config.Auth = []ssh.AuthMethod{ssh.Password(e.servers[i].Password)}
			} else {
				auth, err := e.publicKeyAuthFunc(e.servers[i].IdentityFile)
				if err != nil {
					log.Printf("服务器 %s 读取 IdentityFile 失败：%v", addr, err)
					return
				}
				config.Auth = []ssh.AuthMethod{auth}
			}
			client, err := ssh.Dial("tcp", addr, config)
			if err != nil {
				log.Printf("服务器 %s 建立连接失败：%v", addr, err)
				return
			}
			var rw sshRW
			session, err := client.NewSession()
			if err == nil {
				rw.Stdin, err = session.StdinPipe()
			}
			if err == nil {
				var stdout bytes.Buffer
				session.Stdout = &stdout
				session.Stderr = &stdout
				rw.Stdout = &stdout
			}
			if err == nil {
				err = session.Shell()
			}
			if err != nil {
				log.Printf("服务器 %s 开启 Session 失败：%v", addr, err)
				return
			}
			done := make(chan struct{})
			readCh := make(chan struct{})
			go func() {
				e.readStdout(rw.Stdout, readCh)
				close(done)
			}()
			fmt.Fprintf(rw.Stdin, "\n\necho cook-ssh-executor && date && whoami && echo %s \n", endLabel)
			tm := time.NewTimer(time.Second * 10)
			select {
			case <-done:
				tm.Stop()
			case <-tm.C:
				log.Printf("服务器 %s 建立连接超时", addr)
				close(readCh)
				return
			}
			e.sessions[e.servers[i].Host] = &rw
		}(i)
	}
	wg.Wait()
	log.Printf("%d 个连接已建立", len(e.sessions))
}

func (e *Executor) readStdout(stdout *bytes.Buffer, stdoutClose <-chan struct{}) string {
	var all []byte
	for {
		select {
		case <-stdoutClose:
			return string(all)
		default:
			line, err := stdout.ReadString('\n')
			if strings.TrimSpace(line) == endLabel {
				return string(all)
			}
			all = append(all, []byte(line)...)
			if err != nil && err != io.EOF {
				return string(all)
			}
			time.Sleep(time.Millisecond * 300)
		}
	}
}

func (e *Executor) merge(a, b []*model.Server) []*model.Server {
OUT:
	for i := 0; i < len(b); i++ {
		for j := 0; j < len(a); j++ {
			if reflect.DeepEqual(a[j], b[i]) {
				continue OUT
			}
		}
		a = append(a, b[i])
	}
	return a
}

func (e *Executor) publicKeyAuthFunc(kPath string) (ssh.AuthMethod, error) {
	keyPath, err := homedir.Expand(kPath)
	if err != nil {
		return nil, err
	}
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}
