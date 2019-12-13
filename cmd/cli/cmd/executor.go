package cmd

import (
	"errors"
	"fmt"
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

// Executor ..
type Executor struct {
	conn    []*ssh.Session
	servers []*model.Server
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
				e.connect(strings.Split(args[2], ","), false)
				return
			}
		}
		err = errors.New("参数不完整")
	}

	log.Printf("错误命令：%s %v", s, err)
}

func (e *Executor) connect(tags []string, isAll bool) {
	if isAll {
		for _, server := range dao.Servers {
			e.servers = merge(e.servers, server)
		}
	} else {
		for i := 0; i < len(tags); i++ {
			if ss, has := dao.Servers[tags[i]]; has {
				e.servers = merge(e.servers, ss)
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
				auth, err := publicKeyAuthFunc(e.servers[i].IdentityFile)
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
			session, err := client.NewSession()
			if err != nil {
				log.Printf("服务器 %s 开启 Session 失败：%v", addr, err)
				return
			}
			output, err := session.CombinedOutput("pwd")
			if err != nil {
				log.Printf("服务器 %s 执行失败：%v", addr, err)
				return
			}
			log.Printf("服务器 %s > %s", addr, output)
		}(i)
	}
	wg.Wait()
}

func merge(a, b []*model.Server) []*model.Server {
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

func publicKeyAuthFunc(kPath string) (ssh.AuthMethod, error) {
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
