package plugin

import (
	"bufio"
	"io"
	"net"
	"os"
	"os/exec"
)

type Plugin struct {
	Cmd            *exec.Cmd
	unixSocketName string
}

func NewPlugin(cmdName, socketName string) *Plugin {
	res := new(Plugin)

	res.unixSocketName = "/tmp/" + socketName
	res.Cmd = exec.Command(cmdName, res.unixSocketName)

	// redirect stdout & stderr
	stdout, err := res.Cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	stderr, err := res.Cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	go fwd(stdout, os.Stdout, cmdName)
	go fwd(stderr, os.Stderr, cmdName)

	if err := res.Cmd.Start(); err != nil {
		panic(err)
	}
	return res
}

func (p *Plugin) Dial() (net.Conn, error) {
	return net.Dial("unix", p.unixSocketName)
}

func fwd(in io.Reader, out io.Writer, name string) {
	r := bufio.NewReader(in)
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				out.Write([]byte(name + "> ERROR READING " + err.Error()))
			}
			break
		}
		out.Write([]byte(name + "> " + string(line)))
	}
}
